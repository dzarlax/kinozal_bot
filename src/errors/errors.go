package errors

import "fmt"

// BotError represents errors related to the bot
type BotError struct {
	Message string
	Code    string
	Details map[string]interface{}
}

func (e *BotError) Error() string {
	return fmt.Sprintf("BotError [%s]: %s", e.Code, e.Message)
}

// KinozalError represents errors related to Kinozal
type KinozalError struct {
	Message string
	Details map[string]interface{}
}

func (e *KinozalError) Error() string {
	return fmt.Sprintf("KinozalError: %s", e.Message)
}

// TransmissionError represents errors related to Transmission
type TransmissionError struct {
	Message string
	Details map[string]interface{}
}

func (e *TransmissionError) Error() string {
	return fmt.Sprintf("TransmissionError: %s", e.Message)
}

// Helper functions to create errors
func NewBotError(message, code string, details map[string]interface{}) *BotError {
	return &BotError{
		Message: message,
		Code:    code,
		Details: details,
	}
}

func NewKinozalError(message string, details map[string]interface{}) *KinozalError {
	return &KinozalError{
		Message: message,
		Details: details,
	}
}

func NewTransmissionError(message string, details map[string]interface{}) *TransmissionError {
	return &TransmissionError{
		Message: message,
		Details: details,
	}
}