package http

import "myNotes/core"

type (
	// Request is just a placeholder
	Request struct{}

	// LoginReqest ...
	LoginReqest struct {
		Name, Password string
	}

	// RegisterRequest ...
	RegisterRequest struct {
		Name, Password, Email string
	}

	// VerifyRequest ...
	VerifyRequest struct {
		Name, Password, Code string
	}

	// IDRequest ...
	IDRequest struct {
		ID core.ID
	}

	// OptIDRequest ...
	OptIDRequest struct {
		ID core.ID `urlp:"optional"`
	}

	// ConfigureRequest ...
	ConfigureRequest struct {
		Name, Colors string
	}

	// LikeRequest ...
	LikeRequest struct {
		ID     core.ID
		Target string
		Change bool
	}

	// CommentRequest ...
	CommentRequest struct {
		ID     core.ID
		Target string
	}

	// SaveRequest ...
	SaveRequest struct {
		ID                           core.ID `urlp:"optional"`
		Name, School, Theme, Subject string
		Year, Month                  int
	}

	// PublishRequest ...
	PublishRequest struct {
		ID      core.ID
		Publish bool
	}
)
