package http

import "myNotes/core"

type (
	// LikeResponce ...
	LikeResponce struct {
		Resp  Responce
		Count int
		State bool
	}

	// SearchResponce ...
	SearchResponce struct {
		Resp    Responce
		Results []core.NotePreview
	}

	// AccountResponce ...
	AccountResponce struct {
		Resp    Responce
		Account core.Account
	}

	// ConfigResponce ...
	ConfigResponce struct {
		Resp Responce
		Cfg  core.Config
	}

	// DraftResponce ...
	DraftResponce struct {
		Resp  Responce
		Draft core.Draft
	}

	// SaveResponce ...
	SaveResponce struct {
		Resp Responce
		ID   core.ID
	}

	// NoteResponce ...
	NoteResponce struct {
		Resp Responce
		Note core.Note
	}
)

// Responce is responce sent by RegisterAccount callback
type Responce struct {
	Status string
}

// NResponce creates new responce from error, if error is nil it substitutes success message
func NResponce(err error) Responce {
	status := success
	if err != nil {
		status = err.Error()
	}
	return Responce{status}
}
