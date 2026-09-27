package feedback

import "time"

type Outcome string

const (
	OutcomeShortlist Outcome = "shortlist"
	OutcomeHold      Outcome = "hold"
	OutcomeReject    Outcome = "reject"
)

func (o Outcome) Valid() bool {
	return o == OutcomeShortlist || o == OutcomeHold || o == OutcomeReject
}

type Feedback struct {
	ID              int
	TenantID        string
	SubmissionID    int
	FeedbackByUserID int
	Outcome         Outcome
	ReasonCode      string
	Comments        string
	NextAction      string
	FeedbackAt      time.Time
	CreatedAt       time.Time
}

type CreateInput struct {
	SubmissionID     int
	FeedbackByUserID int
	Outcome          Outcome
	ReasonCode       string
	Comments         string
	NextAction       string
}
