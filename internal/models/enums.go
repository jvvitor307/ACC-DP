package models

type AccFlagType uint32

const (
	FlagNone AccFlagType = iota
	FlagBlue
	FlagYellow
	FlagBlack
	FlagWhite
	FlagCheckered
	FlagPenalty
	FlagGreen
	FlagOrange
)

type AccPenaltyType uint32

const (
	PenaltyNone AccPenaltyType = iota
	PenaltyDriveThrough
	PenaltyStopAndGo10Cutting
	PenaltyStopAndGo20Cutting
	PenaltyStopAndGo30Cutting
	PenaltyDisqualifiedCutting
	PenaltyRemoveBestLapCutting
	PenaltyDriveThroughPitSpeeding
	PenaltyStopAndGo10PitSpeeding
	PenaltyStopAndGo20PitSpeeding
	PenaltyStopAndGo30PitSpeeding
	PenaltyDisqualifiedPitSpeeding
	PenaltyRemoveBestLapPitSpeeding
	PenaltyDisqualifiedIgnoreMandatoryPit
	PenaltyPostRaceTime
	PenaltyDisqualifiedTrolling
	PenaltyDisqualifiedPitEntry
	PenaltyDisqualifiedPitExit
	PenaltyDisqualifiedWrongWay
	PenaltyDriveThroughIgnoredDriverStint
	PenaltyDisqualifiedIgnoredDriverStint
	PenaltyDisqualifiedExceededDriverStintLimit
)

type AccSessionType int32

const (
	SessionUnknown AccSessionType = iota -1
	SessionPractice
	SessionQualify
	SessionRace
	SessionHotlap
	SessionTimeAttack
	SessionDrift
	SessionDrag
	SessionHotstint
	SessionHotstintSuperpole
)

type AccStatus uint32

const (
	StatusOff AccStatus = iota
	StatusReplay
	StatusLive
	StatusPause
)

type AccWheelsType uint32

const (
	WheelsTypeFrontLeft AccWheelsType = iota
	WheelsTypeFrontRight
	WheelsTypeRearLeft
	WheelsTypeRearRight
)

type AccTrackGripStatus uint32

const (
	TrackGripGreen AccTrackGripStatus = iota
	TrackGripFast
	TrackGripOptimum
	TrackGripGreasy
	TrackGripDamp
	TrackGripWet
	TrackGripFlooded
)

type AccRainIntensity uint32

const (
	RainIntensityNone AccRainIntensity = iota
	RainIntensityDrizzle
	RainIntensityLight
	RainIntensityMedium
	RainIntensityHeavy
	RainIntensityStorm
)