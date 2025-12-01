package controller

import "errors"

var (
	ErrParseRequest = errors.New("failed to parse request")

	ErrUserRegister  = errors.New("failed to register user")
	ErrGenerateToken = errors.New("failed to generate token")
	ErrUserLogin     = errors.New("failed to login")

	ErrCreateSession     = errors.New("failed to create an agent session")
	ErrGetSession        = errors.New("failed to get agent sessions")
	ErrDeleteSession     = errors.New("failed to delete an agent session")
	ErrGetSessionMessage = errors.New("failed to get session messages")

	ErrCreateAgent    = errors.New("failed to create an agent")
	ErrCallAgent      = errors.New("error while calling agent")
	ErrSaveAgentSteps = errors.New("failed to save agent steps")

	ErrGetAudioFile     = errors.New("failed to get audio file")
	ErrVoiceRecognition = errors.New("failed to recognize audio")

	ErrGeneratePolicyToken     = errors.New("failed to generate policy token")
	ErrGetKnowledgeMetadata    = errors.New("failed to get knowledge metadata")
	ErrUploadKnowledgeMetadata = errors.New("failed to upload knowledge metadata")
	ErrDeleteKnowledgeMetadata = errors.New("failed to delete knowledge metadata")
)
