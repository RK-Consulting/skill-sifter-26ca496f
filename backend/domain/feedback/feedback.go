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

type ReasonCode string

const (
	ReasonCodeSkillsGap              ReasonCode = "skills_gap"
	ReasonCodeExperienceGap          ReasonCode = "experience_gap"
	ReasonCodeCompensationMismatch   ReasonCode = "compensation_mismatch"
	ReasonCodeLocationMismatch       ReasonCode = "location_mismatch"
	ReasonCodeNoticePeriod           ReasonCode = "notice_period"
	ReasonCodeCandidateNotInterested ReasonCode = "candidate_not_interested"
	ReasonCodeAvailability           ReasonCode = "availability"
	ReasonCodeProfileMismatch        ReasonCode = "profile_mismatch"
	ReasonCodeOther                  ReasonCode = "other"
)

func (r ReasonCode) Valid() bool {
	switch r {
	case ReasonCodeSkillsGap,
		ReasonCodeExperienceGap,
		ReasonCodeCompensationMismatch,
		ReasonCodeLocationMismatch,
		ReasonCodeNoticePeriod,
		ReasonCodeCandidateNotInterested,
		ReasonCodeAvailability,
		ReasonCodeProfileMismatch,
		ReasonCodeOther:
		return true
	default:
		return false
	}
}

type Feedback struct {
	ID                int
	TenantID          string
	SubmissionID      int
	FeedbackByUserID int
	Outcome           Outcome
	ReasonCode        string
	Comments          string
	NextAction        string
	FeedbackAt       time.Time
	CreatedAt        time.Time
}

type CreateInput struct {
	SubmissionID     int
	FeedbackByUserID int
	Outcome          Outcome
	ReasonCode       string
	Comments         string
	NextAction       string
}
