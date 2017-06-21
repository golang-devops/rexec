package comms

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/pkg/errors"
)

func NewExecutorSessions() *executorSessions {
	return &executorSessions{
		sessions: make(map[string]*executorSession),
	}
}

type executorSessions struct {
	sync.RWMutex

	sessions map[string]*executorSession
}

func (e *executorSessions) SessionCount() int {
	return len(e.sessions)
}

func (e *executorSessions) NewSession() *executorSession {
	id := time.Now().Format("2006-01-02 15:04:05.999999999")
	session := &executorSession{
		id:             id,
		writtenCounter: &writtenCounter{},
	}

	e.Lock()
	defer e.Unlock()
	e.sessions[id] = session

	return session
}

func (e *executorSessions) GetSession(id string) (*executorSession, error) {
	session, ok := e.sessions[id]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Session with id %s not found", id))
	}

	return session, nil
}

type executorSession struct {
	id             string
	writtenCounter *writtenCounter
	cmd            *exec.Cmd
	isRunning      bool
	tempFilePath   string
}

func (e *executorSession) handlePanic(logger *log.Entry) {
	if r := recover(); r != nil {
		logger.Error(fmt.Sprintf("PANIC RECOVERED: %v", r))
	}
}

func (e *executorSession) cleanIDForFileName() string {
	s := e.id
	s = strings.Replace(s, " ", "_", -1)
	s = strings.Replace(s, ":", "-", -1)
	s = strings.Replace(s, ".", "-", -1)
	return s
}

func (e *executorSession) SetCommand(cmd *exec.Cmd) {
	e.cmd = cmd
}

func (e *executorSession) StartCommand(logger *log.Entry) error {
	e.isRunning = false

	tempFilePrefix := fmt.Sprintf("rexec-session-%s-", e.cleanIDForFileName())
	tempFile, err := ioutil.TempFile("", tempFilePrefix)
	if err != nil {
		return errors.Wrap(err, "Failed to create temp file")
	}
	e.tempFilePath = tempFile.Name()

	multiWriter := io.MultiWriter(tempFile, e.writtenCounter)
	e.cmd.Stdout = multiWriter
	e.cmd.Stderr = multiWriter

	if err := e.cmd.Start(); err != nil {
		return errors.Wrap(err, "Failed to Start command")
	}
	e.isRunning = true

	go e.wait(logger, tempFile)

	return nil
}

func (e *executorSession) wait(logger *log.Entry, tempFile *os.File) {
	defer e.handlePanic(logger)

	if err := e.cmd.Wait(); err != nil {
		e.isRunning = false
		tempFile.WriteString(fmt.Sprintf("Failed to wait for command, error: %s", err.Error()))

		logger.WithError(err).Error("Failed to wait for command")
		return
	}
	e.isRunning = false

	if err := tempFile.Close(); err != nil {
		logger.WithError(err).Error(fmt.Sprintf("Failed to close file %s", tempFile.Name()))
		return
	}

	//TODO: not removing this since we probably want to read it after process exited
	// if err := os.Remove(tempFile.Name()); err != nil {
	// 	logger.WithError(err).Error(fmt.Sprintf("Failed to delete file %s", tempFile.Name()))
	// 	return
	// }
}

func (e *executorSession) ReadNextLines(offsetLines int) ([]string, error) {
	fileBytes, err := ioutil.ReadFile(e.tempFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read file '%s'", e.tempFilePath)
	}

	allLines := strings.Split(strings.TrimSpace(string(fileBytes)), "\n")

	lines := []string{}
	for _, line := range allLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}

	if offsetLines >= len(lines) && !e.isRunning {
		//return error if client tries to ask for file lines but the lines were all already received and the process is finished running
		return nil, EOF_AND_EXITED
	}

	return lines[offsetLines:], nil
}

type writtenCounter struct {
	c int
}

func (w *writtenCounter) Write(p []byte) (n int, err error) {
	w.c += len(p)
	return len(p), nil
}
