package acc_shm

import "unicode/utf16"

func ToPhysicsPage(raw *PhysicsRawPage) *PhysicsPage {
	if raw == nil {
		return nil
	}

	return &PhysicsPage{
		PacketID:           raw.PacketID,
		Gas:                raw.Gas,
		Brake:              raw.Brake,
		Fuel:               raw.Fuel,
		Gear:               raw.Gear,
		RPM:                raw.RPM,
		SteerAngle:         raw.SteerAngle,
		SpeedKmh:           raw.SpeedKmh,
		Velocity:           raw.Velocity,
		AccG:               raw.AccG,
		WheelSlip:          raw.WheelSlip,
		WheelPressure:      raw.WheelPressure,
		WheelAngularSpeed:  raw.WheelAngularSpeed,
		TyreCoreTemp:       raw.TyreCoreTemp,
		SuspensionTravel:   raw.SuspensionTravel,
		TC:                 raw.TC,
		Heading:            raw.Heading,
		Pitch:              raw.Pitch,
		Roll:               raw.Roll,
		CarDamage:          raw.CarDamage,
		PitLimiterOn:       raw.PitLimiterOn,
		ABS:                raw.ABS,
		AutoShifterOn:      raw.AutoShifterOn,
		TurboBoost:         raw.TurboBoost,
		AirTemp:            raw.AirTemp,
		RoadTemp:           raw.RoadTemp,
		LocalAngularVel:    raw.LocalAngularVel,
		FinalFF:            raw.FinalFF,
		BrakeTemp:          raw.BrakeTemp,
		Clutch:             raw.Clutch,
		IsAIControlled:     raw.IsAIControlled,
		TyreContactPoint:   raw.TyreContactPoint,
		TyreContactNormal:  raw.TyreContactNormal,
		TyreContactHeading: raw.TyreContactHeading,
		BrakeBias:          raw.BrakeBias,
		LocalVelocity:      raw.LocalVelocity,
		SlipRatio:          raw.SlipRatio,
		SlipAngle:          raw.SlipAngle,
		WaterTemp:          raw.WaterTemp,
		BrakePressure:      raw.BrakePressure,
		FrontBrakeCompound: raw.FrontBrakeCompound,
		RearBrakeCompound:  raw.RearBrakeCompound,
		PadLife:            raw.PadLife,
		DiscLife:           raw.DiscLife,
		IgnitionOn:         raw.IgnitionOn,
		StarterEngineOn:    raw.StarterEngineOn,
		IsEngineRunning:    raw.IsEngineRunning,
		KerbVibration:      raw.KerbVibration,
		SlipVibrations:     raw.SlipVibrations,
		GVibrations:        raw.GVibrations,
		ABSVibrations:      raw.ABSVibrations,
	}
}

func ToGraphicsPage(raw *GraphicsRawPage) *GraphicsPage {
	if raw == nil {
		return nil
	}

	return &GraphicsPage{
		PacketID:                 raw.PacketID,
		Status:                   raw.Status,
		Session:                  raw.Session,
		CurrentTime:              decodeUTF16(raw.CurrentTime[:]),
		LastTime:                 decodeUTF16(raw.LastTime[:]),
		BestTime:                 decodeUTF16(raw.BestTime[:]),
		Split:                    decodeUTF16(raw.Split[:]),
		CompletedLaps:            raw.CompletedLaps,
		Position:                 raw.Position,
		ICurrentTime:             raw.ICurrentTime,
		ILastTime:                raw.ILastTime,
		IBestTime:                raw.IBestTime,
		SessionTimeLeft:          raw.SessionTimeLeft,
		DistanceTraveled:         raw.DistanceTraveled,
		IsInPit:                  raw.IsInPit,
		CurrentSectorIndex:       raw.CurrentSectorIndex,
		LastSectorTime:           raw.LastSectorTime,
		NumberOfLaps:             raw.NumberOfLaps,
		TyreCompound:             decodeUTF16(raw.TyreCompound[:]),
		NormalizedCarPosition:    raw.NormalizedCarPosition,
		ActiveCars:               raw.ActiveCars,
		CarCoordinates:           raw.CarCoordinates,
		CarID:                    raw.CarID,
		PlayerCarID:              raw.PlayerCarID,
		PenaltyTime:              raw.PenaltyTime,
		Flag:                     raw.Flag,
		Penalty:                  raw.Penalty,
		IdealLineOn:              raw.IdealLineOn,
		IsInPitLane:              raw.IsInPitLane,
		SurfaceGrip:              raw.SurfaceGrip,
		MandatoryPitDone:         raw.MandatoryPitDone,
		WindSpeed:                raw.WindSpeed,
		WindDirection:            raw.WindDirection,
		IsSetupMenuVisible:       raw.IsSetupMenuVisible,
		MainDisplayIndex:         raw.MainDisplayIndex,
		SecondaryDisplayIndex:    raw.SecondaryDisplayIndex,
		TC:                       raw.TC,
		TCCUT:                    raw.TCCUT,
		EngineMap:                raw.EngineMap,
		ABS:                      raw.ABS,
		FuelXLap:                 raw.FuelXLap,
		RainLights:               raw.RainLights,
		FlashingLights:           raw.FlashingLights,
		LightsStage:              raw.LightsStage,
		ExhaustTemperature:       raw.ExhaustTemperature,
		WiperLV:                  raw.WiperLV,
		DriverStintTotalTimeLeft: raw.DriverStintTotalTimeLeft,
		DriverStintTimeLeft:      raw.DriverStintTimeLeft,
		RainTyres:                raw.RainTyres,
		SessionIndex:             raw.SessionIndex,
		UsedFuel:                 raw.UsedFuel,
		DeltaLapTime:             decodeUTF16(raw.DeltaLapTime[:]),
		IDeltaLapTime:            raw.IDeltaLapTime,
		EstimatedLapTime:         decodeUTF16(raw.EstimatedLapTime[:]),
		IEstimatedLapTime:        raw.IEstimatedLapTime,
		IsDeltaPositive:          raw.IsDeltaPositive,
		ISplit:                   raw.ISplit,
		IsValidLap:               raw.IsValidLap,
		FuelEstimatedLaps:        raw.FuelEstimatedLaps,
		TrackStatus:              decodeUTF16(raw.TrackStatus[:]),
		MissingMandatoryPits:     raw.MissingMandatoryPits,
		Clock:                    raw.Clock,
		DirectionLightsLeft:      raw.DirectionLightsLeft,
		DirectionLightsRight:     raw.DirectionLightsRight,
		GlobalYellow:             raw.GlobalYellow,
		GlobalYellow1:            raw.GlobalYellow1,
		GlobalYellow2:            raw.GlobalYellow2,
		GlobalYellow3:            raw.GlobalYellow3,
		GlobalWhite:              raw.GlobalWhite,
		GlobalGreen:              raw.GlobalGreen,
		GlobalChequered:          raw.GlobalChequered,
		GlobalRed:                raw.GlobalRed,
		MfdTyreSet:               raw.MfdTyreSet,
		MfdFuelToAdd:             raw.MfdFuelToAdd,
		MfdTyrePressureLF:        raw.MfdTyrePressureLF,
		MfdTyrePressureRF:        raw.MfdTyrePressureRF,
		MfdTyrePressureLR:        raw.MfdTyrePressureLR,
		MfdTyrePressureRR:        raw.MfdTyrePressureRR,
		TrackGripStatus:          raw.TrackGripStatus,
		RainIntensity:            raw.RainIntensity,
		RainIntensityIn10Min:     raw.RainIntensityIn10Min,
		RainIntensityIn30Min:     raw.RainIntensityIn30Min,
		CurrentTyreSet:           raw.CurrentTyreSet,
		StrategyTyreSet:          raw.StrategyTyreSet,
		GapAhead:                 raw.GapAhead,
		GapBehind:                raw.GapBehind,
	}
}

func ToStaticPage(raw *StaticRawPage) *StaticPage {
	if raw == nil {
		return nil
	}

	return &StaticPage{
		SMVersion:                decodeUTF16(raw.SMVersion[:]),
		ACVersion:                decodeUTF16(raw.ACVersion[:]),
		NumberOfSessions:         raw.NumberOfSessions,
		NumCars:                  raw.NumCars,
		CarModel:                 decodeUTF16(raw.CarModel[:]),
		Track:                    decodeUTF16(raw.Track[:]),
		PlayerName:               decodeUTF16(raw.PlayerName[:]),
		PlayerSurname:            decodeUTF16(raw.PlayerSurname[:]),
		PlayerNick:               decodeUTF16(raw.PlayerNick[:]),
		SectorCount:              raw.SectorCount,
		MaxRPM:                   raw.MaxRPM,
		MaxFuel:                  raw.MaxFuel,
		PenaltiesEnabled:         raw.PenaltiesEnabled,
		AidFuelRate:              raw.AidFuelRate,
		AidTireRate:              raw.AidTireRate,
		AidMechanicalDamage:      raw.AidMechanicalDamage,
		AllowTyreBlankets:        raw.AllowTyreBlankets,
		AidStability:             raw.AidStability,
		AidAutoclutch:            raw.AidAutoclutch,
		AidAutoBlip:              raw.AidAutoBlip,
		PitWindowStart:           raw.PitWindowStart,
		PitWindowEnd:             raw.PitWindowEnd,
		IsOnline:                 raw.IsOnline,
		DryTyresName:             decodeUTF16(raw.DryTyresName[:]),
		WetTyresName:             decodeUTF16(raw.WetTyresName[:]),
	}
}

func decodeUTF16(raw []uint16) string {
	limit := len(raw)
	for i, code := range raw {
		
		if code == 0 {
			limit = i
			break
		}
	}

	if limit == 0 {
		return ""
	}

	return string(utf16.Decode(raw[:limit]))
}
