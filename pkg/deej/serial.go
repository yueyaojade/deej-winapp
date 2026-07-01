package deej

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"go.uber.org/zap"

	"github.com/yueyaojade/deej-winapp/pkg/deej/util"
)

const (
	// How long without receiving any data before forcing a reconnection
	connectionStaleTimeout = 5 * time.Second

	// Reconnect backoff parameters
	reconnectBaseDelay = 1 * time.Second
	reconnectMaxDelay  = 15 * time.Second
)

// SerialIO provides a deej-aware abstraction layer to managing serial I/O
type SerialIO struct {
	comPort  string
	baudRate uint

	deej   *Deej
	logger *zap.SugaredLogger

	stopChannel chan bool
	connected   bool
	connOptions serial.OpenOptions
	conn        io.ReadWriteCloser

	lastDataTime               time.Time
	lastKnownNumSliders        int
	currentSliderPercentValues []float32

	sliderMoveConsumers []chan SliderMoveEvent

	// protects conn and connected from concurrent access
	closeLock chan struct{}
}

// SliderMoveEvent represents a single slider move captured by deej
type SliderMoveEvent struct {
	SliderID     int
	PercentValue float32
}

var expectedLinePattern = regexp.MustCompile(`^(\d{1,4})(\|(\d{1,4}))*\r?\n$`)

// NewSerialIO creates a SerialIO instance that uses the provided deej
// instance's connection info to establish communications with the arduino chip
func NewSerialIO(deej *Deej, logger *zap.SugaredLogger) (*SerialIO, error) {
	logger = logger.Named("serial")

	sio := &SerialIO{
		deej:                deej,
		logger:              logger,
		stopChannel:         make(chan bool),
		connected:           false,
		conn:                nil,
		sliderMoveConsumers: []chan SliderMoveEvent{},
		lastDataTime:        time.Now(),
		closeLock:           make(chan struct{}, 1),
	}

	logger.Debug("Created serial i/o instance")

	// respond to config changes
	sio.setupOnConfigReload()

	return sio, nil
}

// Start attempts to connect to our arduino chip
func (sio *SerialIO) Start() error {

	// don't allow multiple concurrent connections
	if sio.connected {
		sio.logger.Warn("Already connected, can't start another without closing first")
		return errors.New("serial: connection already active")
	}

	// set minimum read size according to platform (0 for windows, 1 for linux)
	// this prevents a rare bug on windows where serial reads get congested,
	// resulting in significant lag
	minimumReadSize := 0
	if util.Linux() {
		minimumReadSize = 1
	}

	sio.connOptions = serial.OpenOptions{
		PortName:        sio.deej.config.ConnectionInfo.COMPort,
		BaudRate:        uint(sio.deej.config.ConnectionInfo.BaudRate),
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: uint(minimumReadSize),
	}

	if err := sio.doOpen(); err != nil {
		sio.logger.Warnw("Failed to open serial connection", "error", err)
		return fmt.Errorf("open serial connection: %w", err)
	}

	// start the read+reconnect loop
	go sio.readLoop()

	return nil
}

// Stop signals us to shut down our serial connection, if one is active
func (sio *SerialIO) Stop() {
	if sio.connected {
		sio.logger.Debug("Shutting down serial connection")
		sio.stopChannel <- true
	} else {
		sio.logger.Debug("Not currently connected, nothing to stop")
	}
}

// SubscribeToSliderMoveEvents returns an unbuffered channel that receives
// a sliderMoveEvent struct every time a slider moves
func (sio *SerialIO) SubscribeToSliderMoveEvents() chan SliderMoveEvent {
	ch := make(chan SliderMoveEvent)
	sio.sliderMoveConsumers = append(sio.sliderMoveConsumers, ch)

	return ch
}

// doOpen opens the serial port with the configured options
func (sio *SerialIO) doOpen() error {
	conn, err := serial.Open(sio.connOptions)
	if err != nil {
		return err
	}

	sio.conn = conn
	sio.connected = true
	sio.lastDataTime = time.Now()

	namedLogger := sio.logger.Named(strings.ToLower(sio.connOptions.PortName))
	namedLogger.Infow("Connected")

	return nil
}

// readLoop reads slider values from the serial connection and handles
// reconnection when the connection drops or goes stale
func (sio *SerialIO) readLoop() {
	namedLogger := sio.logger.Named(strings.ToLower(sio.connOptions.PortName))

	// channel for read errors (signals reconnect needed)
	readError := make(chan struct{}, 1)

	// start the line reader goroutine
	lineCh := make(chan string, 1)
	go sio.readLines(namedLogger, lineCh, readError)

	// health ticker: checks if we've received data recently
	healthTicker := time.NewTicker(connectionStaleTimeout)
	defer healthTicker.Stop()

	for {
		select {
		case <-sio.stopChannel:
			sio.close(namedLogger)
			return

		case line, ok := <-lineCh:
			if !ok {
				// line channel closed = permanent read failure
				namedLogger.Warn("Serial read terminated, entering reconnect loop")
				sio.close(namedLogger)
				sio.reconnectLoop(namedLogger)
				return
			}

			sio.lastDataTime = time.Now()
			sio.handleLine(namedLogger, line)

		case <-healthTicker.C:
			if sio.connected && time.Since(sio.lastDataTime) > connectionStaleTimeout {
				namedLogger.Warnw("No serial data received for a while, forcing reconnection",
					"idleSeconds", time.Since(sio.lastDataTime).Seconds())
				// close the connection to unblock the reader goroutine
				sio.close(namedLogger)
			}

		case <-readError:
			// reader goroutine reported an error, close and reconnect
			namedLogger.Warn("Read error detected, entering reconnect loop")
			sio.close(namedLogger)
			sio.reconnectLoop(namedLogger)
			return
		}
	}
}

// readLines reads serial data and delivers lines to the provided channel.
// If a permanent read error occurs, it closes lineCh and sends on readError.
func (sio *SerialIO) readLines(logger *zap.SugaredLogger, lineCh chan string, readError chan struct{}) {
	reader := bufio.NewReader(sio.conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			logger.Warnw("Failed to read line from serial", "error", err, "partialLine", line)
			close(lineCh)
			select {
			case readError <- struct{}{}:
			default:
			}
			return
		}

		if sio.deej.Verbose() {
			logger.Debugw("Read new line", "line", line)
		}

		// Deliver line; if the main loop has already exited, just return
		select {
		case lineCh <- line:
		default:
			logger.Debug("readLines: lineCh full, reader may be shutting down")
			return
		}
	}
}

// reconnectLoop keeps trying to reconnect with exponential backoff
// until successful or stopped via stopChannel
func (sio *SerialIO) reconnectLoop(logger *zap.SugaredLogger) {
	delay := reconnectBaseDelay

	for attempt := 1; ; attempt++ {
		// check if we should stop
		select {
		case <-sio.stopChannel:
			logger.Debug("Stopped during reconnect attempt")
			return
		default:
		}

		if sio.connected {
			logger.Debug("Already reconnected, stopping reconnect loop")
			return
		}

		logger.Infow("Attempting serial reconnection", "attempt", attempt, "nextRetry", delay)

		if err := sio.doOpen(); err != nil {
			logger.Warnw("Reconnect attempt failed", "error", err, "retryingIn", delay)
			time.Sleep(delay)
			delay *= 2
			if delay > reconnectMaxDelay {
				delay = reconnectMaxDelay
			}
			continue
		}

		logger.Infow("Successfully reconnected", "attempts", attempt)
		// Start reading on the new connection
		sio.readLoop()
		return
	}
}

// close closes the serial connection if open. Safe to call multiple times.
func (sio *SerialIO) close(logger *zap.SugaredLogger) {
	// Ensure only one goroutine closes at a time
	select {
	case sio.closeLock <- struct{}{}:
	default:
		return // another goroutine is already closing
	}
	defer func() { <-sio.closeLock }()

	if sio.conn != nil {
		if err := sio.conn.Close(); err != nil {
			logger.Warnw("Failed to close serial connection", "error", err)
		} else {
			logger.Debug("Serial connection closed")
		}
	}

	sio.conn = nil
	sio.connected = false
}

func (sio *SerialIO) setupOnConfigReload() {
	configReloadedChannel := sio.deej.config.SubscribeToChanges()

	const stopDelay = 50 * time.Millisecond

	go func() {
		for {
			select {
			case <-configReloadedChannel:

				// make any config reload unset our slider number to ensure process volumes are being re-set
				// (the next read line will emit SliderMoveEvent instances for all sliders)\
				// this needs to happen after a small delay, because the session map will also re-acquire sessions
				// whenever the config file is reloaded, and we don't want it to receive these move events while the map
				// is still cleared. this is kind of ugly, but shouldn't cause any issues
				go func() {
					<-time.After(stopDelay)
					sio.lastKnownNumSliders = 0
				}()

				// if connection params have changed, attempt to stop and start the connection
				if sio.deej.config.ConnectionInfo.COMPort != sio.connOptions.PortName ||
					uint(sio.deej.config.ConnectionInfo.BaudRate) != sio.connOptions.BaudRate {

					sio.logger.Info("Detected change in connection parameters, attempting to renew connection")
					sio.Stop()

					// let the connection close
					<-time.After(stopDelay)

					if err := sio.Start(); err != nil {
						sio.logger.Warnw("Failed to renew connection after parameter change", "error", err)
					} else {
						sio.logger.Debug("Renewed connection successfully")
					}
				}
			}
		}
	}()
}

func (sio *SerialIO) handleLine(logger *zap.SugaredLogger, line string) {

	// this function receives an unsanitized line which is guaranteed to end with LF,
	// but most lines will end with CRLF. it may also have garbage instead of
	// deej-formatted values, so we must check for that! just ignore bad ones
	if !expectedLinePattern.MatchString(line) {
		return
	}

	// trim the suffix (accept both CRLF and LF)
	line = strings.TrimRight(line, "\r\n")

	// split on pipe (|), this gives a slice of numerical strings between "0" and "1023"
	splitLine := strings.Split(line, "|")
	numSliders := len(splitLine)

	// update our slider count, if needed - this will send slider move events for all
	if numSliders != sio.lastKnownNumSliders {
		logger.Infow("Detected sliders", "amount", numSliders)
		sio.lastKnownNumSliders = numSliders
		sio.currentSliderPercentValues = make([]float32, numSliders)

		// reset everything to be an impossible value to force the slider move event later
		for idx := range sio.currentSliderPercentValues {
			sio.currentSliderPercentValues[idx] = -1.0
		}
	}

	// for each slider:
	moveEvents := []SliderMoveEvent{}
	for sliderIdx, stringValue := range splitLine {

		// convert string values to integers ("1023" -> 1023)
		number, _ := strconv.Atoi(stringValue)

		// turns out the first line could come out dirty sometimes (i.e. "4558|925|41|643|220")
		// so let's check the first number for correctness just in case
		if sliderIdx == 0 && number > 1023 {
			sio.logger.Debugw("Got malformed line from serial, ignoring", "line", line)
			return
		}

		// map the value from raw to a "dirty" float between 0 and 1 (e.g. 0.15451...)
		dirtyFloat := float32(number) / 1023.0

		// normalize it to an actual volume scalar between 0.0 and 1.0 with 2 points of precision
		normalizedScalar := util.NormalizeScalar(dirtyFloat)

		// if sliders are inverted, take the complement of 1.0
		if sio.deej.config.InvertSliders {
			normalizedScalar = 1 - normalizedScalar
		}

		// check if it changes the desired state (could just be a jumpy raw slider value)
		if util.SignificantlyDifferent(sio.currentSliderPercentValues[sliderIdx], normalizedScalar, sio.deej.config.NoiseReductionLevel) {

			// if it does, update the saved value and create a move event
			sio.currentSliderPercentValues[sliderIdx] = normalizedScalar

			moveEvents = append(moveEvents, SliderMoveEvent{
				SliderID:     sliderIdx,
				PercentValue: normalizedScalar,
			})

			if sio.deej.Verbose() {
				logger.Debugw("Slider moved", "event", moveEvents[len(moveEvents)-1])
			}
		}
	}

	// deliver move events if there are any, towards all potential consumers
	if len(moveEvents) > 0 {
		for _, consumer := range sio.sliderMoveConsumers {
			for _, moveEvent := range moveEvents {
				consumer <- moveEvent
			}
		}
	}
}
