package acc_shm

const (
	PhysicsSharedMemoryName  = "Local\\acpmf_physics"
	GraphicsSharedMemoryName = "Local\\acpmf_graphics"
	StaticSharedMemoryName   = "Local\\acpmf_static"
)

const (
	PhysicsPageSize  = 800
	GraphicsPageSize = 1588
	StaticPageSize   = 820
)

type ACCFlagType int32

const (
	ACCNoFlag        ACCFlagType = 0
	ACCBlueFlag      ACCFlagType = 1
	ACCYellowFlag    ACCFlagType = 2
	ACCBlackFlag     ACCFlagType = 3
	ACCWhiteFlag     ACCFlagType = 4
	ACCCheckeredFlag ACCFlagType = 5
	ACCPenaltyFlag   ACCFlagType = 6
	ACCGreenFlag     ACCFlagType = 7
	ACCOrangeFlag    ACCFlagType = 8
)

type ACCPenaltyType int32

const (
	ACCPenaltyNone                                 ACCPenaltyType = 0
	ACCPenaltyDriveThroughCutting                  ACCPenaltyType = 1
	ACCPenaltyStopAndGo10Cutting                   ACCPenaltyType = 2
	ACCPenaltyStopAndGo20Cutting                   ACCPenaltyType = 3
	ACCPenaltyStopAndGo30Cutting                   ACCPenaltyType = 4
	ACCPenaltyDisqualifiedCutting                  ACCPenaltyType = 5
	ACCPenaltyRemoveBestLapTimeCutting             ACCPenaltyType = 6
	ACCPenaltyDriveThroughPitSpeeding              ACCPenaltyType = 7
	ACCPenaltyStopAndGo10PitSpeeding               ACCPenaltyType = 8
	ACCPenaltyStopAndGo20PitSpeeding               ACCPenaltyType = 9
	ACCPenaltyStopAndGo30PitSpeeding               ACCPenaltyType = 10
	ACCPenaltyDisqualifiedPitSpeeding              ACCPenaltyType = 11
	ACCPenaltyRemoveBestLapTimePitSpeeding         ACCPenaltyType = 12
	ACCPenaltyDisqualifiedIgnoredMandatoryPit      ACCPenaltyType = 13
	ACCPenaltyPostRaceTime                         ACCPenaltyType = 14
	ACCPenaltyDisqualifiedTrolling                 ACCPenaltyType = 15
	ACCPenaltyDisqualifiedPitEntry                 ACCPenaltyType = 16
	ACCPenaltyDisqualifiedPitExit                  ACCPenaltyType = 17
	ACCPenaltyDisqualifiedWrongWay                 ACCPenaltyType = 18
	ACCPenaltyDriveThroughIgnoredDriverStint       ACCPenaltyType = 19
	ACCPenaltyDisqualifiedIgnoredDriverStint       ACCPenaltyType = 20
	ACCPenaltyDisqualifiedExceededDriverStintLimit ACCPenaltyType = 21
)

type ACCSessionType int32

const (
	ACCSessionUnknown           ACCSessionType = -1
	ACCSessionPractice          ACCSessionType = 0
	ACCSessionQualify           ACCSessionType = 1
	ACCSessionRace              ACCSessionType = 2
	ACCSessionHotlap            ACCSessionType = 3
	ACCSessionTimeAttack        ACCSessionType = 4
	ACCSessionDrift             ACCSessionType = 5
	ACCSessionDrag              ACCSessionType = 6
	ACCSessionHotstint          ACCSessionType = 7
	ACCSessionHotstintSuperpole ACCSessionType = 8
)

type ACCStatus int32

const (
	ACCStatusOff    ACCStatus = 0
	ACCStatusReplay ACCStatus = 1
	ACCStatusLive   ACCStatus = 2
	ACCStatusPause  ACCStatus = 3
)

type ACCTrackGripStatus int32

const (
	ACCTrackGripGreen   ACCTrackGripStatus = 0
	ACCTrackGripFast    ACCTrackGripStatus = 1
	ACCTrackGripOptimum ACCTrackGripStatus = 2
	ACCTrackGripGreasy  ACCTrackGripStatus = 3
	ACCTrackGripDamp    ACCTrackGripStatus = 4
	ACCTrackGripWet     ACCTrackGripStatus = 5
	ACCTrackGripFlooded ACCTrackGripStatus = 6
)

type ACCRainIntensity int32

const (
	ACCRainNone         ACCRainIntensity = 0
	ACCRainDrizzle      ACCRainIntensity = 1
	ACCRainLight        ACCRainIntensity = 2
	ACCRainMedium       ACCRainIntensity = 3
	ACCRainHeavy        ACCRainIntensity = 4
	ACCRainThunderstorm ACCRainIntensity = 5
)

type PhysicsRawPage struct {
	PacketID           int32
	Gas                float32
	Brake              float32
	Fuel               float32
	Gear               int32
	RPM                int32
	SteerAngle         float32
	SpeedKmh           float32
	Velocity           [3]float32
	AccG               [3]float32
	WheelSlip          [4]float32
	WheelLoad          [4]float32
	WheelPressure      [4]float32
	WheelAngularSpeed  [4]float32
	TyreWear           [4]float32
	TyreDirtyLevel     [4]float32
	TyreCoreTemp       [4]float32
	CamberRad          [4]float32
	SuspensionTravel   [4]float32
	DRS                float32
	TC                 float32
	Heading            float32
	Pitch              float32
	Roll               float32
	CGHeight           float32
	CarDamage          [5]float32
	NumberOfTyresOut   int32
	PitLimiterOn       int32
	ABS                float32
	KERSCharge         float32
	KERSInput          float32
	AutoShifterOn      int32
	RideHeight         [2]float32
	TurboBoost         float32
	Ballast            float32
	AirDensity         float32
	AirTemp            float32
	RoadTemp           float32
	LocalAngularVel    [3]float32
	FinalFF            float32
	PerformanceMeter   float32
	EngineBrake        int32
	ERSRecoveryLevel   int32
	ERSPowerLevel      int32
	ERSHeatCharging    int32
	ERSIsCharging      int32
	KERSCurrentKJ      float32
	DRSAvailable       int32
	DRSEnabled         int32
	BrakeTemp          [4]float32
	Clutch             float32
	TyreTempI          [4]float32
	TyreTempM          [4]float32
	TyreTempO          [4]float32
	IsAIControlled     int32
	TyreContactPoint   [4][3]float32
	TyreContactNormal  [4][3]float32
	TyreContactHeading [4][3]float32
	BrakeBias          float32
	LocalVelocity      [3]float32
	P2PActivation      int32
	P2PStatus          int32
	CurrentMaxRPM      float32
	Mz                 [4]float32
	Fx                 [4]float32
	Fy                 [4]float32
	SlipRatio          [4]float32
	SlipAngle          [4]float32
	TCInAction         int32
	ABSInAction        int32
	SuspensionDamage   [4]float32
	TyreTemp           [4]float32
	WaterTemp          float32
	BrakePressure      [4]float32
	FrontBrakeCompound int32
	RearBrakeCompound  int32
	PadLife            [4]float32
	DiscLife           [4]float32
	IgnitionOn         int32
	StarterEngineOn    int32
	IsEngineRunning    int32
	KerbVibration      float32
	SlipVibrations     float32
	GVibrations        float32
	ABSVibrations      float32
}

type GraphicsRawPage struct {
	PacketID                 int32
	Status                   ACCStatus
	Session                  ACCSessionType
	CurrentTime              [15]uint16
	LastTime                 [15]uint16
	BestTime                 [15]uint16
	Split                    [15]uint16
	CompletedLaps            int32
	Position                 int32
	ICurrentTime             int32
	ILastTime                int32
	IBestTime                int32
	SessionTimeLeft          float32
	DistanceTraveled         float32
	IsInPit                  int32
	CurrentSectorIndex       int32
	LastSectorTime           int32
	NumberOfLaps             int32
	TyreCompound             [34]uint16
	ReplayTimeMultiplier     float32
	NormalizedCarPosition    float32
	ActiveCars               int32
	CarCoordinates           [60][3]float32
	CarID                    [60]int32
	PlayerCarID              int32
	PenaltyTime              float32
	Flag                     ACCFlagType
	Penalty                  ACCPenaltyType
	IdealLineOn              int32
	IsInPitLane              int32
	SurfaceGrip              float32
	MandatoryPitDone         int32
	WindSpeed                float32
	WindDirection            float32
	IsSetupMenuVisible       int32
	MainDisplayIndex         int32
	SecondaryDisplayIndex    int32
	TC                       int32
	TCCUT                    int32
	EngineMap                int32
	ABS                      int32
	FuelXLap                 float32
	RainLights               int32
	FlashingLights           int32
	LightsStage              int32
	ExhaustTemperature       float32
	WiperLV                  int32
	DriverStintTotalTimeLeft int32
	DriverStintTimeLeft      int32
	RainTyres                int32
	SessionIndex             int32
	UsedFuel                 float32
	DeltaLapTime             [16]uint16
	IDeltaLapTime            int32
	EstimatedLapTime         [16]uint16
	IEstimatedLapTime        int32
	IsDeltaPositive          int32
	ISplit                   int32
	IsValidLap               int32
	FuelEstimatedLaps        float32
	TrackStatus              [34]uint16
	MissingMandatoryPits     int32
	Clock                    float32
	DirectionLightsLeft      int32
	DirectionLightsRight     int32
	GlobalYellow             int32
	GlobalYellow1            int32
	GlobalYellow2            int32
	GlobalYellow3            int32
	GlobalWhite              int32
	GlobalGreen              int32
	GlobalChequered          int32
	GlobalRed                int32
	MfdTyreSet               int32
	MfdFuelToAdd             float32
	MfdTyrePressureLF        float32
	MfdTyrePressureRF        float32
	MfdTyrePressureLR        float32
	MfdTyrePressureRR        float32
	TrackGripStatus          ACCTrackGripStatus
	RainIntensity            ACCRainIntensity
	RainIntensityIn10Min     ACCRainIntensity
	RainIntensityIn30Min     ACCRainIntensity
	CurrentTyreSet           int32
	StrategyTyreSet          int32
	GapAhead                 int32
	GapBehind                int32
}

type StaticRawPage struct {
	SMVersion                [15]uint16
	ACVersion                [15]uint16
	NumberOfSessions         int32
	NumCars                  int32
	CarModel                 [33]uint16
	Track                    [33]uint16
	PlayerName               [33]uint16
	PlayerSurname            [33]uint16
	PlayerNick               [34]uint16
	SectorCount              int32
	MaxTorque                float32
	MaxPower                 float32
	MaxRPM                   int32
	MaxFuel                  float32
	SuspensionMaxTravel      [4]float32
	TyreRadius               [4]float32
	MaxTurboBoost            float32
	Deprecated1              float32
	Deprecated2              float32
	PenaltiesEnabled         int32
	AidFuelRate              float32
	AidTireRate              float32
	AidMechanicalDamage      float32
	AllowTyreBlankets        float32
	AidStability             float32
	AidAutoclutch            int32
	AidAutoBlip              int32
	HasDRS                   int32
	HasERS                   int32
	HasKERS                  int32
	KERSMaxJ                 float32
	EngineBrakeSettingsCount int32
	ERSPowerControllerCount  int32
	TrackSplineLength        float32
	TrackConfiguration       [34]uint16
	ErsMaxJ                  float32
	IsTimedRace              int32
	HasExtraLap              int32
	CarSkin                  [34]uint16
	ReversedGridPositions    int32
	PitWindowStart           int32
	PitWindowEnd             int32
	IsOnline                 int32
	DryTyresName             [33]uint16
	WetTyresName             [33]uint16
}

// PhysicsPage mirrors PhysicsRawPage semantics. It remains numeric-only.
type PhysicsPage struct {
	PacketID           int32
	Gas                float32
	Brake              float32
	Fuel               float32
	Gear               int32
	RPM                int32
	SteerAngle         float32
	SpeedKmh           float32
	Velocity           [3]float32
	AccG               [3]float32
	WheelSlip          [4]float32
	WheelPressure      [4]float32
	WheelAngularSpeed  [4]float32
	TyreCoreTemp       [4]float32
	SuspensionTravel   [4]float32
	TC                 float32
	Heading            float32
	Pitch              float32
	Roll               float32
	CarDamage          [5]float32
	PitLimiterOn       int32
	ABS                float32
	AutoShifterOn      int32
	TurboBoost         float32
	AirTemp            float32
	RoadTemp           float32
	LocalAngularVel    [3]float32
	FinalFF            float32
	BrakeTemp          [4]float32
	Clutch             float32
	IsAIControlled     int32
	TyreContactPoint   [4][3]float32
	TyreContactNormal  [4][3]float32
	TyreContactHeading [4][3]float32
	BrakeBias          float32
	LocalVelocity      [3]float32
	SlipRatio          [4]float32
	SlipAngle          [4]float32
	WaterTemp          float32
	BrakePressure      [4]float32
	FrontBrakeCompound int32
	RearBrakeCompound  int32
	PadLife            [4]float32
	DiscLife           [4]float32
	IgnitionOn         int32
	StarterEngineOn    int32
	IsEngineRunning    int32
	KerbVibration      float32
	SlipVibrations     float32
	GVibrations        float32
	ABSVibrations      float32
}

type GraphicsPage struct {
	PacketID                 int32
	Status                   ACCStatus
	Session                  ACCSessionType
	CurrentTime              string
	LastTime                 string
	BestTime                 string
	Split                    string
	CompletedLaps            int32
	Position                 int32
	ICurrentTime             int32
	ILastTime                int32
	IBestTime                int32
	SessionTimeLeft          float32
	DistanceTraveled         float32
	IsInPit                  int32
	CurrentSectorIndex       int32
	LastSectorTime           int32
	NumberOfLaps             int32
	TyreCompound             string
	NormalizedCarPosition    float32
	ActiveCars               int32
	CarCoordinates           [60][3]float32
	CarID                    [60]int32
	PlayerCarID              int32
	PenaltyTime              float32
	Flag                     ACCFlagType
	Penalty                  ACCPenaltyType
	IdealLineOn              int32
	IsInPitLane              int32
	SurfaceGrip              float32
	MandatoryPitDone         int32
	WindSpeed                float32
	WindDirection            float32
	IsSetupMenuVisible       int32
	MainDisplayIndex         int32
	SecondaryDisplayIndex    int32
	TC                       int32
	TCCUT                    int32
	EngineMap                int32
	ABS                      int32
	FuelXLap                 float32
	RainLights               int32
	FlashingLights           int32
	LightsStage              int32
	ExhaustTemperature       float32
	WiperLV                  int32
	DriverStintTotalTimeLeft int32
	DriverStintTimeLeft      int32
	RainTyres                int32
	SessionIndex             int32
	UsedFuel                 float32
	DeltaLapTime             string
	IDeltaLapTime            int32
	EstimatedLapTime         string
	IEstimatedLapTime        int32
	IsDeltaPositive          int32
	ISplit                   int32
	IsValidLap               int32
	FuelEstimatedLaps        float32
	TrackStatus              string
	MissingMandatoryPits     int32
	Clock                    float32
	DirectionLightsLeft      int32
	DirectionLightsRight     int32
	GlobalYellow             int32
	GlobalYellow1            int32
	GlobalYellow2            int32
	GlobalYellow3            int32
	GlobalWhite              int32
	GlobalGreen              int32
	GlobalChequered          int32
	GlobalRed                int32
	MfdTyreSet               int32
	MfdFuelToAdd             float32
	MfdTyrePressureLF        float32
	MfdTyrePressureRF        float32
	MfdTyrePressureLR        float32
	MfdTyrePressureRR        float32
	TrackGripStatus          ACCTrackGripStatus
	RainIntensity            ACCRainIntensity
	RainIntensityIn10Min     ACCRainIntensity
	RainIntensityIn30Min     ACCRainIntensity
	CurrentTyreSet           int32
	StrategyTyreSet          int32
	GapAhead                 int32
	GapBehind                int32
}

type StaticPage struct {
	SMVersion                string
	ACVersion                string
	NumberOfSessions         int32
	NumCars                  int32
	CarModel                 string
	Track                    string
	PlayerName               string
	PlayerSurname            string
	PlayerNick               string
	SectorCount              int32
	MaxRPM                   int32
	MaxFuel                  float32
	PenaltiesEnabled         int32
	AidFuelRate              float32
	AidTireRate              float32
	AidMechanicalDamage      float32
	AllowTyreBlankets        float32
	AidStability             float32
	AidAutoclutch            int32
	AidAutoBlip              int32
	PitWindowStart           int32
	PitWindowEnd             int32
	IsOnline                 int32
	DryTyresName             string
	WetTyresName             string
}
