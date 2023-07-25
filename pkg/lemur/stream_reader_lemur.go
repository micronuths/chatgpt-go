package lemur

import (
	"bufio"
	"bytes"
	utils "chatgpt-go/pkg/lemur/internal"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type streamReaderLemur struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator utils.ErrorAccumulator
	unmarshaler    utils.Unmarshaler
}
type LemurResponseST struct {
	Origin string `json:"origin"`
	Data   string `json:"data"`
	Code   int    `json:"code"`
}
type LemurResponseSEC struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int                      `json:"created"`
	Model   string                   `json:"model"`
	Choices []LemurResponseSecChoice `json:"choices"`
}
type LemurResponseSecChoice struct {
	Index        int                         `json:"index"`
	Delta        LemurResponseSecChoiceDelta `json:"delta"`
	FinishReason any                         `json:"finish_reason"`
}

type LemurResponseSecChoiceDelta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var (
	headerDataLemur  = []byte("data: ")
	errorPrefixLemur = []byte(`data: {"error":`)
)

func (stream *streamReaderLemur) Recv() (response LemurResponseSEC, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	response, err = stream.processLines()
	return
}

//nolint:gocognit
func (stream *streamReaderLemur) processLines() (LemurResponseSEC, error) {
	var (
		emptyMessagesCount uint
		hasErrorPrefix     bool
	)

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')

		if readErr != nil || hasErrorPrefix {
			respErr := stream.unmarshalError()
			if respErr != nil {
				return *new(LemurResponseSEC), fmt.Errorf("respErr.Error, %w", respErr.Error)
			}
			return *new(LemurResponseSEC), fmt.Errorf("readErr, %w", readErr)
		}

		noSpaceLine := bytes.TrimSpace(rawLine)
		if bytes.HasPrefix(noSpaceLine, errorPrefixLemur) {
			hasErrorPrefix = true
		}
		if !bytes.HasPrefix(noSpaceLine, headerDataLemur) || hasErrorPrefix {
			if hasErrorPrefix {
				noSpaceLine = bytes.TrimPrefix(noSpaceLine, headerData)
			}
			writeErr := stream.errAccumulator.Write(noSpaceLine)
			if writeErr != nil {
				return *new(LemurResponseSEC), fmt.Errorf("error, %w", writeErr)
			}
			emptyMessagesCount++
			if emptyMessagesCount > stream.emptyMessagesLimit {
				return *new(LemurResponseSEC), ErrTooManyEmptyStreamMessages
			}

			continue
		}

		noPrefixLine := bytes.TrimPrefix(noSpaceLine, headerDataLemur)
		if string(noPrefixLine) == "[DONE]" {
			stream.isFinished = true
			return *new(LemurResponseSEC), io.EOF
		}

		var responseST LemurResponseST

		unmarshalErr := stream.unmarshaler.Unmarshal(noPrefixLine, &responseST)
		if unmarshalErr != nil {
			return *new(LemurResponseSEC), unmarshalErr
		}

		// if responseST.Code != 0 {
		// 	return *new(LemurResponseSEC), fmt.Errorf(" responseST.Code=%d", responseST.Code)
		// }
		responseSec := LemurResponseSEC{
			Choices: []LemurResponseSecChoice{
				{
					Delta: LemurResponseSecChoiceDelta{
						Content: "",
					},
				},
			},
		}
		for _, v := range strings.Split(responseST.Data, "\n\n") {
			if len(v) == 0 {
				continue
			}
			var responseSecTmp LemurResponseSEC
			noPrefixLine02 := bytes.TrimPrefix([]byte(v), headerDataLemur)
			err := stream.unmarshaler.Unmarshal(noPrefixLine02, &responseSecTmp)
			if err != nil {
				continue
			}
			responseSec.ID = responseSecTmp.ID
			responseSec.Object = responseSecTmp.Object
			responseSec.Created = responseSecTmp.Created
			responseSec.Model = responseSecTmp.Model

			if len(responseSecTmp.Choices) > 0 {
				for _, v := range responseSecTmp.Choices {
					responseSec.Choices[0].Delta.Content = responseSec.Choices[0].Delta.Content + v.Delta.Content
				}
			}

		}

		// fmt.Println("stream_reader_lemur.go/responseST=", responseSec)

		return responseSec, nil
	}
}

func (stream *streamReaderLemur) unmarshalError() (errResp *ErrorResponse) {
	errBytes := stream.errAccumulator.Bytes()
	if len(errBytes) == 0 {
		return
	}

	err := stream.unmarshaler.Unmarshal(errBytes, &errResp)
	if err != nil {
		errResp = nil
	}

	return
}

func (stream *streamReaderLemur) Close() {
	stream.response.Body.Close()
}
