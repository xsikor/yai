package engine

import (
	"context"
	"errors"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"time"
)

type EngineOutput struct {
	content string
	last    bool
}

func (d EngineOutput) IsLast() bool {
	return d.last
}

func (d EngineOutput) GetContent() string {
	return d.content
}

type Engine struct {
	client   *openai.Client
	messages []openai.ChatCompletionMessage
	channel  chan EngineOutput
	mode     EngineMode
	running  bool
}

func NewEngine(mode EngineMode) *Engine {
	return &Engine{
		client:   openai.NewClient("xxx"),
		messages: make([]openai.ChatCompletionMessage, 0),
		channel:  make(chan EngineOutput),
		mode:     mode,
		running:  false,
	}
}

func (e *Engine) Channel() chan EngineOutput {
	return e.channel
}

func (e *Engine) Interrupt() *Engine {
	e.channel <- EngineOutput{
		content: "\n\nInterrupt !",
		last:    true,
	}

	e.running = false

	return e
}

func (e *Engine) Reset() *Engine {
	e.messages = []openai.ChatCompletionMessage{}

	return e
}

func (e *Engine) SetMode(mode EngineMode) *Engine {
	e.mode = mode

	return e
}

func (e *Engine) GetMode() EngineMode {
	return e.mode
}

func (e *Engine) StreamChatCompletion(input string) error {

	ctx := context.Background()

	e.running = true

	e.appendUserMessage(input)

	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 1000,
		Messages:  e.prepareCompletionMessages(),
		Stream:    true,
	}

	stream, err := e.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Printf("error on stream creation: %v", err)
		return err
	}
	defer stream.Close()

	var output string

	for {
		if e.running {
			resp, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				e.channel <- EngineOutput{
					content: "",
					last:    true,
				}
				e.running = false
				e.appendAssistantMessage(output)

				return nil
			}

			if err != nil {
				log.Printf("error on stream read: %v", err)
				e.running = false
				return err
			}

			delta := resp.Choices[0].Delta.Content

			output += delta

			e.channel <- EngineOutput{
				content: delta,
				last:    false,
			}

			time.Sleep(time.Millisecond * 1)
		} else {
			stream.Close()

			return nil
		}
	}
}

func (e *Engine) appendUserMessage(content string) *Engine {
	e.messages = append(e.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})

	return e
}

func (e *Engine) appendAssistantMessage(content string) *Engine {
	e.messages = append(e.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: content,
	})

	return e
}

func (e *Engine) prepareCompletionMessages() []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: e.prepareSystemMessageContent(),
		},
	}
	for _, m := range e.messages {
		messages = append(messages, m)
	}

	return messages
}

func (e *Engine) prepareSystemMessageContent() string {
	prompt := "You are Yo, an helpful AI command line assistant running in a terminal, created by Jonathan VUILLEMIN (github.com/ekkinox). "

	switch e.mode {
	case ChatEngineMode:
		prompt += "You will provide an answer for my input the most helpful possible, rendered in markdown format. "
	case RunEngineMode:
		prompt += "You will always return ONLY a single command line that fulfills my input, without any explanation or descriptive text. "
		prompt += "This command line cannot have new lines, use instead separators like && and ;. "
	}

	prompt += "My operating system is linux, my distribution is Fedora release 37 (Thirty Seven), my home directory is /home/jonathan, my shell is zsh."

	return prompt
}
