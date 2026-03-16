# ACC Shared Memory v1.8.12

- Field is working as intended
- Field is not used by ACC --> **_[NOT USED]_**
- New / reworked entry(/ies) since last build
- \* indicates double fields

### Order of wheels:

| Etiqueta | Descrição   |
| :------- | :---------- |
| FL       | Front Left  |
| FR       | Front Right |
| RL       | Rear Left   |
| RR       | Rear Right  |

## SPageFilePhysics

The following members change at each graphic step. They all refer to the player’s car.

### Dicionário de Dados: Telemetria (Shared Memory / UDP)

| Variável (Tipo)                  | Descrição                                                          |
| :------------------------------- | :----------------------------------------------------------------- |
| `int packetId`                   | Current step index                                                 |
| `float gas`                      | Gas pedal input value (from -0 to 1.0)                             |
| `float brake`                    | Brake pedal input value (from -0 to 1.0)                           |
| `float fuel`                     | Amount of fuel remaining in kg                                     |
| `int gear`                       | Current gear                                                       |
| `int rpm`                        | Engine revolutions per minute                                      |
| `float steerAngle`               | Steering input value (from -1.0 to 1.0)                            |
| `float speedKmh`                 | Car speed in km/h                                                  |
| `float[3] velocity`              | Car velocity vector in global coordinates                          |
| `float[3] accG`                  | Car acceleration vector in global coordinates                      |
| `float[4] wheelSlip`             | Tyre slip for each tyre [FL, FR, RL, RR]                           |
| `float[4] wheelLoad`             | **_[NOT USED]_** Wheel load for each tyre [FL, FR, RL, RR]         |
| `float[4] wheelPressure`         | Tyre pressure [FL, FR, RL, RR]                                     |
| `float[4] wheelAngularSpeed`     | Wheel angular speed in rad/s [FL, FR, RL, RR]                      |
| `float[4] tyreWear`              | **_[NOT USED]_** Tyre wear [FL, FR, RL, RR]                        |
| `float[4] tyreDirtyLevel`        | **_[NOT USED]_** Dirt accumulated on tyre surface [FL, FR, RL, RR] |
| `float[4] TyreCoreTemp`          | \* Tyre rubber core temperature [FL, FR, RL, RR]                   |
| `float[4] camberRAD`             | **_[NOT USED]_** Wheels camber in radians [FL, FR, RL, RR]         |
| `float[4] suspensionTravel`      | Suspension travel [FL, FR, RL, RR]                                 |
| `float drs`                      | **_[NOT USED]_** DRS on                                            |
| `float tc`                       | \*\* TC in action                                                  |
| `float heading`                  | Car yaw orientation                                                |
| `float pitch`                    | Car pitch orientation                                              |
| `float roll`                     | Car roll orientation                                               |
| `float cgHeight`                 | **_[NOT USED]_** Centre of gravity height                          |
| `float[5] carDamage`             | Car damage: front 0, rear 1, left 2, right 3, centre 4             |
| `int numberOfTyresOut`           | **_[NOT USED]_** Number of tyres out of track                      |
| `int pitLimiterOn`               | Pit limiter is on                                                  |
| `float abs`                      | \*\*\* ABS in action                                               |
| `float kersCharge`               | **_[NOT USED]_** Not used in ACC                                   |
| `float kersInput`                | **_[NOT USED]_** Not used in ACC                                   |
| `int autoshifterOn`              | Automatic transmission on                                          |
| `float[2] rideHeight`            | **_[NOT USED]_** Ride height: 0 front, 1 rear                      |
| `float turboBoost`               | Car turbo level                                                    |
| `float ballast`                  | **_[NOT USED]_** Car ballast in kg / Not implemented               |
| `float airDensity`               | **_[NOT USED]_** Air density                                       |
| `float airTemp`                  | Air temperature                                                    |
| `float roadTemp`                 | Road temperature                                                   |
| `float[3] localAngularVel`       | Car angular velocity vector in local coordinates                   |
| `float finalFF`                  | Force feedback signal                                              |
| `float performanceMeter`         | **_[NOT USED]_** Not used in ACC                                   |
| `int engineBrake`                | **_[NOT USED]_** Not used in ACC                                   |
| `int ersRecoveryLevel`           | **_[NOT USED]_** Not used in ACC                                   |
| `int ersPowerLevel`              | **_[NOT USED]_** Not used in ACC                                   |
| `int ersHeatCharging`            | **_[NOT USED]_** Not used in ACC                                   |
| `int ersIsCharging`              | **_[NOT USED]_** Not used in ACC                                   |
| `float kersCurrentKJ`            | **_[NOT USED]_** Not used in ACC                                   |
| `int drsAvailable`               | **_[NOT USED]_** Not used in ACC                                   |
| `int drsEnabled`                 | **_[NOT USED]_** Not used in ACC                                   |
| `float[4] brakeTemp`             | Brake discs temperatures                                           |
| `float clutch`                   | Clutch pedal input value (from -0 to 1.0)                          |
| `float[4] tyreTempI`             | **_[NOT USED]_** Not shown in ACC                                  |
| `float[4] tyreTempM`             | **_[NOT USED]_** Not shown in ACC                                  |
| `float[4] tyreTempO`             | **_[NOT USED]_** Not shown in ACC                                  |
| `int isAIControlled`             | Car is controlled by the AI                                        |
| `float[4][3] tyreContactPoint`   | Tyre contact point global coordinates [FL, FR, RL, RR] [x,y,z]     |
| `float[4][3] tyreContactNormal`  | Tyre contact normal [FL, FR, RL, RR] [x,y,z]                       |
| `float[4][3] tyreContactHeading` | Tyre contact heading [FL, FR, RL, RR] [x,y,z]                      |
| `float brakeBias`                | Front brake bias, see Appendix 4                                   |
| `float[3] localVelocity`         | Car velocity vector in local coordinates                           |
| `int P2PActivation`              | **_[NOT USED]_** Not used in ACC                                   |
| `int P2PStatus`                  | **_[NOT USED]_** Not used in ACC                                   |
| `float currentMaxRpm`            | **_[NOT USED]_** Maximum engine rpm                                |
| `float[4] mz`                    | **_[NOT USED]_** Not shown in ACC                                  |
| `float[4] fx`                    | **_[NOT USED]_** Not shown in ACC                                  |
| `float[4] fy`                    | **_[NOT USED]_** Not shown in ACC                                  |
| `float[4] slipRatio`             | Tyre slip ratio [FL, FR, RL, RR] in radians                        |
| `float[4] slipAngle`             | Tyre slip angle [FL, FR, RL, RR]                                   |
| `int tcinAction`                 | **_[NOT USED]_** \*\* TC in action                                 |
| `int absInAction`                | **_[NOT USED]_** \*\*\* ABS in action                              |
| `float[4] suspensionDamage`      | **_[NOT USED]_** Suspensions damage levels [FL, FR, RL, RR]        |
| `float[4] tyreTemp`              | **_[NOT USED]_** \* Tyres core temperatures [FL, FR, RL, RR]       |
| `float waterTemp`                | Water Temperature                                                  |
| `float[4] brakePressure`         | Brake pressure [FL, FR, RL, RR] see Appendix 2                     |
| `int frontBrakeCompound`         | Brake pad compund front                                            |
| `int rearBrakeCompound`          | Brake pad compund rear                                             |
| `float[4] padLife`               | Brake pad wear [FL, FR, RL, RR]                                    |
| `float[4] discLife`              | Brake disk wear [FL, FR, RL, RR]                                   |
| `int ignitionOn`                 | Ignition switch set to on?                                         |
| `int starterEngineOn`            | Starter Switch set to on?                                          |
| `int isEngineRunning`            | Engine running?                                                    |
| `float kerbVibration`            | Vibrations sent to the FFB, could be used for motion rigs          |
| `float slipVibrations`           | Vibrations sent to the FFB, could be used for motion rigs          |
| `float gVibrations`              | Vibrations sent to the FFB, could be used for motion rigs          |
| `float absVibrations`            | Vibrations sent to the FFB, could be used for motion rigs          |

## SPageFileGraphic

The following members are updated at each graphical step. They mostly refer to player’s car
except for carCoordinates and carID, which refer to the cars currently on track.

### Dicionário de Dados: Telemetria (Graphics / Session Status)

| Variável (Tipo)                           | Descrição                                              |
| :---------------------------------------- | :----------------------------------------------------- |
| `int packetId`                            | Current step index                                     |
| `ACC_STATUS status`                       | See enums ACC_STATUS                                   |
| `ACC_SESSION_TYPE session`                | See enums ACC_SESSION_TYPE                             |
| `wchar_t[15] currentTime`                 | Current lap time in wide character                     |
| `wchar_t[15] lastTime`                    | Last lap time in wide character                        |
| `wchar_t[15] bestTime`                    | Best lap time in wide character                        |
| `wchar_t[15] split`                       | Last split time in wide character                      |
| `int completedLaps`                       | \* No of completed laps                                |
| `int position`                            | Current player position                                |
| `int iCurrentTime`                        | Current lap time in milliseconds                       |
| `int iLastTime`                           | Last lap time in milliseconds                          |
| `int iBestTime`                           | Best lap time in milliseconds                          |
| `float sessionTimeLeft`                   | Session time left                                      |
| `float distanceTraveled`                  | Distance travelled in the current stint                |
| `int isInPit`                             | Car is pitting                                         |
| `int currentSectorIndex`                  | Current track sector                                   |
| `int lastSectorTime`                      | Last sector time in milliseconds                       |
| `int numberOfLaps`                        | \* Number of completed laps                            |
| `wchar_t[33] tyreCompound`                | Tyre compound used                                     |
| `float replayTimeMultiplier`              | **_[NOT USED]_** Not used in ACC                       |
| `float normalizedCarPosition`             | Car position on track spline (0.0 start to 1.0 finish) |
| `int activeCars`                          | Number of cars on track                                |
| `float[60][3] carCoordinates`             | Coordinates of cars on track                           |
| `int[60] carID`                           | Car IDs of cars on track                               |
| `int playerCarID`                         | Player Car ID                                          |
| `float penaltyTime`                       | Penalty time to wait                                   |
| `ACC_FLAG_TYPE flag`                      | See enums ACC_FLAG_TYPE                                |
| `ACC_PENALTY_TYPE penalty`                | See enums ACC_PENALTY_TYPE                             |
| `int idealLineOn`                         | Ideal line on                                          |
| `int isInPitLane`                         | Car is in pit lane                                     |
| `float surfaceGrip`                       | Ideal line friction coefficient                        |
| `int mandatoryPitDone`                    | Mandatory pit is completed                             |
| `float windSpeed`                         | Wind speed in m/s                                      |
| `float windDirection`                     | wind direction in radians                              |
| `int isSetupMenuVisible`                  | Car is working on setup                                |
| `int mainDisplayIndex`                    | current car main display index, see Appendix 1         |
| `int secondaryDisplyIndex`                | current car secondary display index                    |
| `int TC`                                  | Traction control level                                 |
| `int TCCUT`                               | Traction control cut level                             |
| `int EngineMap`                           | Current engine map                                     |
| `int ABS`                                 | ABS level                                              |
| `float fuelXLap`                          | Average fuel consumed per lap in liters                |
| `int rainLights`                          | Rain lights on                                         |
| `int flashingLights`                      | Flashing lights on                                     |
| `int lightsStage`                         | Current lights stage                                   |
| `float exhaustTemperature`                | Exhaust temperature                                    |
| `int wiperLV`                             | Current wiper stage                                    |
| `int driverStintTotalTimeLeft`            | Time the driver is allowed to drive/race (ms)          |
| `int driverStintTimeLeft`                 | Time the driver is allowed to drive/stint (ms)         |
| `int rainTyres`                           | Are rain tyres equipped                                |
| `int sessionIndex`                        | Session index                                          |
| `float usedFuel`                          | Used fuel since last time refueling                    |
| `wchar_t[15] deltaLapTime`                | Delta time in wide character                           |
| `int iDeltaLapTime`                       | Delta time in milliseconds                             |
| `wchar_t[15] estimatedLapTime`            | Estimated lap time in wide character                   |
| `int iEstimatedLapTime`                   | Estimated lap time in milliseconds                     |
| `int isDeltaPositive`                     | Delta positive (1) or negative (0)                     |
| `int iSplit`                              | Last split time in milliseconds                        |
| `int isValidLap`                          | Check if Lap is valid for timing                       |
| `float fuelEstimatedLaps`                 | Laps possible with current fuel level                  |
| `wchar_t[33] trackStatus`                 | Status of track                                        |
| `int missingMandatoryPits`                | Mandatory pitstops the player still has to do          |
| `float Clock`                             | Time of day in seconds                                 |
| `int directionLightsLeft`                 | Is Blinker left on                                     |
| `int directionLightsRight`                | Is Blinker right on                                    |
| `int GlobalYellow`                        | Yellow Flag is out?                                    |
| `int GlobalYellow1`                       | Yellow Flag in Sector 1 is out?                        |
| `int GlobalYellow2`                       | Yellow Flag in Sector 2 is out?                        |
| `int GlobalYellow3`                       | Yellow Flag in Sector 3 is out?                        |
| `int GlobalWhite`                         | White Flag is out?                                     |
| `int GlobalGreen`                         | Green Flag is out?                                     |
| `int GlobalChequered`                     | Checkered Flag is out?                                 |
| `int GlobalRed`                           | Red Flag is out?                                       |
| `int mfdTyreSet`                          | # of tyre set on the MFD                               |
| `float mfdFuelToAdd`                      | How much fuel to add on the MFD                        |
| `float mfdTyrePressureLF`                 | Tyre pressure left front on the MFD                    |
| `float mfdTyrePressureRF`                 | Tyre pressure right front on the MFD                   |
| `float mfdTyrePressureLR`                 | Tyre pressure left rear on the MFD                     |
| `float mfdTyrePressureRR`                 | Tyre pressure right rear on the MFD                    |
| `ACC_TRACK_GRIP_STATUS trackGripStatus`   | See enums ACC_TRACK_GRIP_STATUS                        |
| `ACC_RAIN_INTENSITY rainIntensity`        | See enums ACC_RAIN_INTENSITY                           |
| `ACC_RAIN_INTENSITY rainIntensityIn10min` | See enums ACC_RAIN_INTENSITY                           |
| `ACC_RAIN_INTENSITY rainIntensityIn30min` | See enums ACC_RAIN_INTENSITY                           |
| `int currentTyreSet`                      | Tyre Set currently in use                              |
| `int strategyTyreSet`                     | Next Tyre set per strategy                             |
| `int gapAhead`                            | Distance in ms to car in front                         |
| `int gapBehind`                           | Distance in ms to car behind                           |

## SPageFileStatic

The following members are initialized when the instance starts and never changes until the
instance is closed.

### Dicionário de Dados: Telemetria (Static Info)

| Variável (Tipo)                | Descrição                                 |
| :----------------------------- | :---------------------------------------- |
| `wchar_t[15] smVersion`        | Shared memory version                     |
| `wchar_t[15] acVersion`        | Assetto Corsa version                     |
| `int numberOfSessions`         | Number of sessions                        |
| `int numCars`                  | Number of cars                            |
| `wchar_t[33] carModel`         | Player car model                          |
| `wchar_t[33] track`            | Track name                                |
| `wchar_t[33] playerName`       | Player name                               |
| `wchar_t[33] playerSurname`    | Player surname                            |
| `wchar_t[33] playerNick`       | Player nickname                           |
| `int sectorCount`              | Number of sectors                         |
| `float maxTorque`              | **_[NOT USED]_** Not shown in ACC         |
| `float maxPower`               | **_[NOT USED]_** Not shown in ACC         |
| `int maxRpm`                   | Maximum rpm                               |
| `float maxFuel`                | Maximum fuel tank capacity                |
| `float[4] suspensionMaxTravel` | **_[NOT USED]_** Not shown in ACC         |
| `float[4] tyreRadius`          | **_[NOT USED]_** Not shown in ACC         |
| `float maxTurboBoost`          | **_[NOT USED]_** Maximum turbo boost      |
| `float deprecated_1`           | **_[NOT USED]_** Padding / Deprecated     |
| `float deprecated_2`           | **_[NOT USED]_** Padding / Deprecated     |
| `int penaltiesEnabled`         | Penalties enabled                         |
| `float aidFuelRate`            | Fuel consumption rate                     |
| `float aidTireRate`            | Tyre wear rate                            |
| `float aidMechanicalDamage`    | Mechanical damage rate                    |
| `float AllowTyreBlankets`      | Not allowed in Blancpain endurance series |
| `float aidStability`           | Stability control used                    |
| `int aidAutoclutch`            | Auto clutch used                          |
| `int aidAutoBlip`              | Always true in ACC                        |
| `int hasDRS`                   | **_[NOT USED]_** Not used in ACC          |
| `int hasERS`                   | **_[NOT USED]_** Not used in ACC          |
| `int hasKERS`                  | **_[NOT USED]_** Not used in ACC          |
| `float kersMaxJ`               | **_[NOT USED]_** Not used in ACC          |
| `int engineBrakeSettingsCount` | **_[NOT USED]_** Not used in ACC          |
| `int ersPowerControllerCount`  | **_[NOT USED]_** Not used in ACC          |
| `float trackSplineLength`      | **_[NOT USED]_** Not used in ACC          |
| `wchar_t trackConfiguration`   | **_[NOT USED]_** Not used in ACC          |
| `float ersMaxJ`                | **_[NOT USED]_** Not used in ACC          |
| `int isTimedRace`              | **_[NOT USED]_** Not used in ACC          |
| `int hasExtraLap`              | **_[NOT USED]_** Not used in ACC          |
| `wchar_t[33] carSkin`          | **_[NOT USED]_** Not used in ACC          |
| `int reversedGridPositions`    | **_[NOT USED]_** Not used in ACC          |
| `int PitWindowStart`           | Pit window opening time                   |
| `int PitWindowEnd`             | Pit windows closing time                  |
| `int isOnline`                 | If is a multiplayer session               |
| `wchar_t[33] dryTyresName`     | Name of the dry tyres                     |
| `wchar_t[33] wetTyresName`     | Name of the wet tyres                     |

## Enums

### ACC_FLAG_TYPE

| Constante              | Valor | Descrição                                    |
| :--------------------- | :---: | :------------------------------------------- |
| **ACC_NO_FLAG**        |   0   | Nenhuma bandeira ativa                       |
| **ACC_BLUE_FLAG**      |   1   | Bandeira Azul (ceder passagem)               |
| **ACC_YELLOW_FLAG**    |   2   | Bandeira Amarela (perigo no setor)           |
| **ACC_BLACK_FLAG**     |   3   | Bandeira Preta (desclassificação)            |
| **ACC_WHITE_FLAG**     |   4   | Bandeira Branca (última volta / carro lento) |
| **ACC_CHECKERED_FLAG** |   5   | Bandeira Quadriculada (fim de sessão)        |
| **ACC_PENALTY_FLAG**   |   6   | Bandeira de Penalidade                       |
| **ACC_GREEN_FLAG**     |   7   | Bandeira Verde (pista liberada)              |
| **ACC_ORANGE_FLAG**    |   8   | Bandeira Laranja (carro com danos técnicos)  |

### ACC_PENALTY_TYPE

| Constante                                     | Valor | Causa / Descrição                              |
| :-------------------------------------------- | :---: | :--------------------------------------------- |
| **ACC_None**                                  |   0   | Nenhuma penalidade                             |
| **ACC_DriveThrough_Cutting**                  |   1   | Drive Through (Corte de pista)                 |
| **ACC_StopAndGo_10_Cutting**                  |   2   | Stop & Go 10s (Corte de pista)                 |
| **ACC_StopAndGo_20_Cutting**                  |   3   | Stop & Go 20s (Corte de pista)                 |
| **ACC_StopAndGo_30_Cutting**                  |   4   | Stop & Go 30s (Corte de pista)                 |
| **ACC_Disqualified_Cutting**                  |   5   | Desclassificado (Corte de pista)               |
| **ACC_RemoveBestLaptime_Cutting**             |   6   | Remoção da melhor volta (Corte de pista)       |
| **ACC_DriveThrough_PitSpeeding**              |   7   | Drive Through (Excesso de velocidade no Pit)   |
| **ACC_StopAndGo_10_PitSpeeding**              |   8   | Stop & Go 10s (Excesso de velocidade no Pit)   |
| **ACC_StopAndGo_20_PitSpeeding**              |   9   | Stop & Go 20s (Excesso de velocidade no Pit)   |
| **ACC_StopAndGo_30_PitSpeeding**              |  10   | Stop & Go 30s (Excesso de velocidade no Pit)   |
| **ACC_Disqualified_PitSpeeding**              |  11   | Desclassificado (Excesso de velocidade no Pit) |
| **ACC_RemoveBestLaptime_PitSpeeding**         |  12   | Remoção da melhor volta (Excesso no Pit)       |
| **ACC_Disqualified_IgnoredMandatoryPit**      |  13   | Desclassificado (Ignorou Pit obrigatório)      |
| **ACC_PostRaceTime**                          |  14   | Penalidade de tempo pós-corrida                |
| **ACC_Disqualified_Trolling**                 |  15   | Desclassificado (Comportamento antidesportivo) |
| **ACC_Disqualified_PitEntry**                 |  16   | Desclassificado (Entrada irregular nos Pits)   |
| **ACC_Disqualified_PitExit**                  |  17   | Desclassificado (Saída irregular nos Pits)     |
| **ACC_Disqualified_Wrongway**                 |  18   | Desclassificado (Sentido contrário)            |
| **ACC_DriveThrough_IgnoredDriverStint**       |  19   | Drive Through (Stint de piloto ignorado)       |
| **ACC_Disqualified_IgnoredDriverStint**       |  20   | Desclassificado (Stint de piloto ignorado)     |
| **ACC_Disqualified_ExceededDriverStintLimit** |  21   | Desclassificado (Limite de Stint excedido)     |

### ACC_SESSION_TYPE

| Constante                 | Valor | Descrição               |
| :------------------------ | :---: | :---------------------- |
| **ACC_UNKNOWN**           |  -1   | Sessão Desconhecida     |
| **ACC_PRACTICE**          |   0   | Treino Livre (Practice) |
| **ACC_QUALIFY**           |   1   | Qualificação            |
| **ACC_RACE**              |   2   | Corrida                 |
| **ACC_HOTLAP**            |   3   | Hotlap                  |
| **ACC_TIMEATTACK**        |   4   | Time Attack             |
| **ACC_DRIFT**             |   5   | Drift                   |
| **ACC_DRAG**              |   6   | Drag                    |
| **ACC_HOTSTINT**          |   7   | Hotstint                |
| **ACC_HOTSTINTSUPERPOLE** |   8   | Hotstint Superpole      |

### ACC_STATUS

| Constante      | Valor | Descrição                  |
| :------------- | :---: | :------------------------- |
| **ACC_OFF**    |   0   | Simulador desligado / Menu |
| **ACC_REPLAY** |   1   | Em modo Replay             |
| **ACC_LIVE**   |   2   | Em tempo real (Pista)      |
| **ACC_PAUSE**  |   3   | Jogo pausado               |

### ACC_WHEELS_TYPE

| Constante          | Valor | Posição da Roda         |
| :----------------- | :---: | :---------------------- |
| **ACC_FrontLeft**  |   0   | Dianteira Esquerda (FL) |
| **ACC_FrontRight** |   1   | Dianteira Direita (FR)  |
| **ACC_RearLeft**   |   2   | Traseira Esquerda (RL)  |
| **ACC_RearRight**  |   3   | Traseira Direita (RR)   |

### ACC_TRACK_GRIP_STATUS

| Constante       | Valor | Estado da Pista                       |
| :-------------- | :---: | :------------------------------------ |
| **ACC_GREEN**   |   0   | Verde (Baixa aderência inicial)       |
| **ACC_FAST**    |   1   | Rápida (Aderência melhorando)         |
| **ACC_OPTIMUM** |   2   | Ideal (Máxima aderência)              |
| **ACC_GREASY**  |   3   | Escorregadia (Início de chuva / óleo) |
| **ACC_DAMP**    |   4   | Úmida                                 |
| **ACC_WET**     |   5   | Molhada                               |
| **ACC_FLOODED** |   6   | Alagada                               |

### ACC_RAIN_INTENSITY

| Constante            | Valor | Nível de Precipitação            |
| :------------------- | :---: | :------------------------------- |
| **ACC_NO_RAIN**      |   0   | Sem chuva                        |
| **ACC_DRIZZLE**      |   1   | Garoa                            |
| **ACC_LIGHT_RAIN**   |   2   | Chuva Leve                       |
| **ACC_MEDIUM_RAIN**  |   3   | Chuva Média                      |
| **ACC_HEAVY_RAIN**   |   4   | Chuva Pesada                     |
| **ACC_THUNDERSTORM** |   5   | Tempestade / Tempestade Elétrica |

## Appendix 1 – mainDisplayIndex

### Tabela de Disponibilidade: Modelos GT3 (2018)

| Veículo                            | Ano  | Page 1 | Page 2 | Page 3 | Page 4 |
| :--------------------------------- | :--: | :----: | :----: | :----: | :----: |
| **Aston Martin Vantage V12 GT3**   | 2013 |   0    |   1    |   -    |   -    |
| **Audi R8 LMS**                    | 2015 |   0    |   1    |   2    |   3    |
| **Bentley Continental GT3**        | 2015 |   0    |   1    |   -    |   -    |
| **Bentley Continental GT3**        | 2018 |   0    |   1    |   2    |   -    |
| **BMW M6 GT3**                     | 2017 |   0    |   -    |   -    |   -    |
| **Emil Frey Jaguar G3**            | 2012 |   0    |   1    |   -    |   -    |
| **Ferrari 488 GT3**                | 2018 |   0    |   1    |   2    |   -    |
| **Honda NSX GT3**                  | 2017 |   0    |   -    |   -    |   -    |
| **Lamborghini Gallardo G3 Reiter** | 2017 |   0    |   1    |   -    |   -    |
| **Lamborghini Huracan GT3**        | 2015 |   4    |   0    |   2    |   3    |
| **Lamborghini Huracan ST**         | 2015 |   0    |   -    |   -    |   -    |
| **Lexus RCF GT3**                  | 2016 |   0    |   -    |   -    |   -    |
| **McLaren 650S GT3**               | 2015 |   0    |   1    |   2    |   4    |
| **Mercedes AMG GT3**               | 2015 |   0    |   1    |   -    |   -    |
| **Nissan GTR Nismo GT3**           | 2015 |   1    |   3    |   4    |   0    |
| **Nissan GTR Nismo GT3**           | 2018 |   1    |   3    |   4    |   0    |
| **Porsche 991 GT3 R**              | 2018 |   0    |   1    |   2    |   3    |
| **Porsche 991 II GT3 Cup**         | 2017 |   0    |   1    |   2    |   3    |

### Tabela de Disponibilidade: Modelos GT3 (2019)

| Veículo                         | Ano  | Page 1 | Page 2 | Page 3 | Page 4 |
| :------------------------------ | :--: | :----: | :----: | :----: | :----: |
| **Aston Martin V8 Vantage GT3** | 2019 |   0    |   -    |   -    |   -    |
| **Audi R8 LMS Evo**             | 2019 |   1    |   2    |   3    |   0    |
| **Honda NSX GT3 Evo**           | 2019 |   0    |   -    |   -    |   -    |
| **Lamborghini Huracan GT3 EVO** | 2019 |   4    |   0    |   2    |   3    |
| **McLaren 720S GT3**            | 2019 |   0    |   1    |   2    |   3    |
| **Porsche 911 II GT3 R**        | 2019 |   1    |   2    |   3    |   0    |

### Tabela de Disponibilidade: Modelos GT4

| Veículo                          | Ano  | Page 1 | Page 2 | Page 3 | Page 4 |
| :------------------------------- | :--: | :----: | :----: | :----: | :----: |
| **Alpine A110 GT4**              | 2018 |   0    |   -    |   -    |   -    |
| **Aston Martin Vantage AMR GT4** | 2018 |   0    |   1    |   -    |   -    |
| **Audi R8 LMS GT4**              | 2016 |   0    |   1    |   -    |   -    |
| **BMW M4 GT4**                   | 2018 |   0    |   1    |   -    |   -    |
| **Chevrolet Camaro GT4 R**       | 2017 |   2    |   4    |   -    |   -    |
| **Ginetta G55 GT4**              | 2012 |   0    |   4    |   -    |   -    |
| **Ktm Xbow GT4**                 | 2016 |   0    |   1    |   3    |   4    |
| **Maserati Gran Turismo MC GT4** | 2016 |   0    |   -    |   -    |   -    |
| **McLaren 570s GT4**             | 2016 |   0    |   -    |   -    |   -    |
| **Mercedes AMG GT4**             | 2016 |   0    |   1    |   -    |   -    |
| **Porsche 718 Cayman GT4 MR**    | 2019 |   0    |   -    |   -    |   -    |

### Tabela de Disponibilidade: Modelos GT3 (2020)

| Veículo                  | Ano  | Page 1 | Page 2 | Page 3 | Page 4 |
| :----------------------- | :--: | :----: | :----: | :----: | :----: |
| **Ferrari 488 GT3 Evo**  | 2020 |   0    |   1    |   2    |   -    |
| **Mercedes AMG GT3 Evo** | 2020 |   0    |   1    |   -    |   -    |

| Veículo        | Ano  | Page 1 | Page 2 | Page 3 | Page 4 |
| :------------- | :--: | :----: | :----: | :----: | :----: |
| **BMW M4 GT3** | 2021 |   0    |   1    |   -    |   -    |

| Veículo                         | Ano  | Page 1 | Page 2 | Page 3 | Page 4 |
| :------------------------------ | :--: | :----: | :----: | :----: | :----: |
| **Audi R8 LMS Evo II**          | 2022 |   0    |   1    |   2    |   -    |
| **BMW M2 Cup**                  | 2020 |   0    |   1    |   -    |   -    |
| **Ferrari 488 Challenge Evo**   | 2020 |   0    |   1    |   2    |   -    |
| **Lamborghini Huracan ST Evo2** | 2021 |   0    |   -    |   -    |   -    |
| **Porsche 992 GT3 Cup**         | 2021 |   0    |   1    |   3    |   -    |

## Appendix 2 – carModel

### Tabela de Referência: GT3 (2018)

| Nome do Veículo                    | Ano  | Kunos ID                       |
| :--------------------------------- | :--: | :----------------------------- |
| **Aston Martin Vantage V12 GT3**   | 2013 | `amr_v12_vantage_gt3`          |
| **Audi R8 LMS**                    | 2015 | `audi_r8_lms`                  |
| **Bentley Continental GT3**        | 2015 | `bentley_continental_gt3_2016` |
| **Bentley Continental GT3**        | 2018 | `bentley_continental_gt3_2018` |
| **BMW M6 GT3**                     | 2017 | `bmw_m6_gt3`                   |
| **Emil Frey Jaguar G3**            | 2012 | `jaguar_g3`                    |
| **Ferrari 488 GT3**                | 2018 | `ferrari_488_gt3`              |
| **Honda NSX GT3**                  | 2017 | `honda_nsx_gt3`                |
| **Lamborghini Gallardo G3 Reiter** | 2017 | `lamborghini_gallardo_rex`     |
| **Lamborghini Huracan GT3**        | 2015 | `lamborghini_huracan_gt3`      |
| **Lamborghini Huracan ST**         | 2015 | `lamborghini_huracan_st`       |
| **Lexus RCF GT3**                  | 2016 | `lexus_rc_f_gt3`               |
| **McLaren 650S GT3**               | 2015 | `mclaren_650s_gt3`             |
| **Mercedes AMG GT3**               | 2015 | `mercedes_amg_gt3`             |
| **Nissan GTR Nismo GT3**           | 2015 | `nissan_gt_r_gt3_2017`         |
| **Nissan GTR Nismo GT3**           | 2018 | `nissan_gt_r_gt3_2018`         |
| **Porsche 991 GT3 R**              | 2018 | `porsche_991_gt3_r`            |
| **Porsche 991 II GT3 Cup**         | 2017 | `porsche_991ii_gt3_cup`        |

### Tabela de Referência: GT3 (2019)

| Nome do Veículo                 | Ano  | Kunos ID                      |
| :------------------------------ | :--: | :---------------------------- |
| **Aston Martin V8 Vantage GT3** | 2019 | `amr_v8_vantage_gt3`          |
| **Audi R8 LMS Evo**             | 2019 | `audi_r8_lms_evo`             |
| **Honda NSX GT3 Evo**           | 2019 | `honda_nsx_gt3_evo`           |
| **Lamborghini Huracan GT3 EVO** | 2019 | `lamborghini_huracan_gt3_evo` |
| **McLaren 720S GT3**            | 2019 | `mclaren_720s_gt3`            |
| **Porsche 911 II GT3 R**        | 2019 | `porsche_991ii_gt3_r`         |

### Tabela de Referência: GT4

| Nome do Veículo                  | Ano  | Kunos ID                    |
| :------------------------------- | :--: | :-------------------------- |
| **Alpine A110 GT4**              | 2018 | `alpine_a110_gt4`           |
| **Aston Martin Vantage AMR GT4** | 2018 | `amr_v8_vantage_gt4`        |
| **Audi R8 LMS GT4**              | 2016 | `audi_r8_gt4`               |
| **BMW M4 GT4**                   | 2018 | `bmw_m4_gt4`                |
| **Chevrolet Camaro GT4 R**       | 2017 | `chevrolet_camaro_gt4r`     |
| **Ginetta G55 GT4**              | 2012 | `ginetta_g55_gt4`           |
| **Ktm Xbow GT4**                 | 2016 | `ktm_xbow_gt4`              |
| **Maserati Gran Turismo MC GT4** | 2016 | `maserati_mc_gt4`           |
| **McLaren 570s GT4**             | 2016 | `mclaren_570s_gt4`          |
| **Mercedes AMG GT4**             | 2016 | `mercedes_amg_gt4`          |
| **Porsche 718 Cayman GT4 MR**    | 2019 | `porsche_718_cayman_gt4_mr` |

### Tabela de Referência: GT3 (2020)

| Nome do Veículo          | Ano  | Kunos ID               |
| :----------------------- | :--: | :--------------------- |
| **Ferrari 488 GT3 Evo**  | 2020 | `ferrari_488_gt3_evo`  |
| **Mercedes AMG GT3 Evo** | 2020 | `mercedes_amg_gt3_evo` |

### Tabela de Referência: GT3 (2021)

| Nome do Veículo | Ano  | Kunos ID     |
| :-------------- | :--: | :----------- |
| **BMW M4 vGT3** | 2021 | `bmw_m4_gt3` |

### Tabela de Referência: Challengers Pack (2022)

| Nome do Veículo                 | Ano  | Kunos ID                      |
| :------------------------------ | :--: | :---------------------------- |
| **Audi R8 LMS Evo II**          | 2022 | `audi_r8_lms_evo_ii`          |
| **BMW M2 Cup**                  | 2020 | `bmw_m2_cs_racing`            |
| **Ferrari 488 Challenge Evo**   | 2020 | `ferrari_488_challenge_evo`   |
| **Lamborghini Huracan ST Evo2** | 2021 | `lamborghini_huracan_st_evo2` |
| **Porsche 992 GT3 Cup**         | 2021 | `porsche_992_gt3_cup`         |

## Appendix 3 – brakePressure

### GT3 - 2018 Rear engine

| Nome do Veículo                    | Ano  | Coeficiente Dianteiro | Coeficiente Traseiro |
| :--------------------------------- | :--: | :-------------------: | :------------------: |
| **Aston Martin Vantage V12 GT3**   | 2013 |        7.9585         |        7.9585        |
| **Audi R8 LMS**                    | 2015 |        7.5980         |        7.4855        |
| **Bentley Continental GT3**        | 2015 |        7.9585         |        7.9585        |
| **Bentley Continental GT3**        | 2018 |        7.9585         |        7.9585        |
| **BMW M6 GT3**                     | 2017 |        7.9585         |        7.9585        |
| **Emil Frey Jaguar G3**            | 2012 |        7.9585         |        7.9585        |
| **Ferrari 488 GT3**                | 2018 |        7.5980         |        7.4855        |
| **Honda NSX GT3**                  | 2017 |        7.5980         |        7.4855        |
| **Lamborghini Gallardo G3 Reiter** | 2017 |        7.5980         |        7.4855        |
| **Lamborghini Huracan GT3**        | 2015 |        7.5980         |        7.4855        |
| **Lamborghini Huracan ST**         | 2015 |        7.5980         |        7.4855        |
| **Lexus RCF GT3**                  | 2016 |        7.9585         |        7.9585        |
| **McLaren 650S GT3**               | 2015 |        7.5980         |        7.4855        |
| **Mercedes AMG GT3**               | 2015 |        7.9585         |        7.9585        |
| **Nissan GTR Nismo GT3**           | 2015 |        7.9585         |        7.9585        |
| **Nissan GTR Nismo GT3**           | 2018 |        7.9585         |        7.9585        |
| **Porsche 991 GT3 R**              | 2018 |        7.1497         |        6.7715        |
| **Porsche 991 II GT3 Cup**         | 2017 |        7.1497         |        6.7715        |

### GT3 – 2019 Rear engine

| Nome do Veículo                 | Ano  | Coeficiente Dianteiro | Coeficiente Traseiro |
| :------------------------------ | :--: | :-------------------: | :------------------: |
| **Aston Martin V8 Vantage GT3** | 2019 |        7.9585         |        7.9585        |
| **Audi R8 LMS Evo**             | 2019 |        7.5980         |        7.4855        |
| **Honda NSX GT3 Evo**           | 2019 |        7.5980         |        7.4855        |
| **Lamborghini Huracan GT3 EVO** | 2019 |        7.5980         |        7.4855        |
| **McLaren 720S GT3**            | 2019 |        7.5980         |        7.4855        |
| **Porsche 911 II GT3 R**        | 2019 |        7.1497         |        6.7715        |

### GT4 Rear engine

| Nome do Veículo                  | Ano  | Coeficiente Dianteiro | Coeficiente Traseiro |
| :------------------------------- | :--: | :-------------------: | :------------------: |
| **Alpine A110 GT4**              | 2018 |        10.0000        |       10.0000        |
| **Aston Martin Vantage AMR GT4** | 2018 |        10.0000        |       10.0000        |
| **Audi R8 LMS GT4**              | 2016 |        10.0000        |       10.0000        |
| **BMW M4 GT4**                   | 2018 |        7.2886         |       10.0000        |
| **Chevrolet Camaro GT4 R**       | 2017 |        10.0000        |       10.0000        |
| **Ginetta G55 GT4**              | 2012 |        10.0000        |       10.0000        |
| **Ktm Xbow GT4**                 | 2016 |        10.0000        |       10.0000        |
| **Maserati Gran Turismo MC GT4** | 2016 |        7.7768         |        7.6142        |
| **McLaren 570s GT4**             | 2016 |        10.0000        |       10.0000        |
| **Mercedes AMG GT4**             | 2016 |        10.0000        |       10.0000        |
| **Porsche 718 Cayman GT4 MR**    | 2019 |        10.0000        |       10.0000        |

### GT3 – 2020 Rear engine

| Nome do Veículo          | Ano  | Coeficiente Dianteiro | Coeficiente Traseiro |
| :----------------------- | :--: | :-------------------: | :------------------: |
| **Ferrari 488 GT3 Evo**  | 2020 |        7.5980         |        7.4855        |
| **Mercedes AMG GT3 Evo** | 2020 |        7.9585         |        7.9585        |

### GT3 – 2021 Rear engine

| Nome do Veículo | Ano  | Coeficiente Dianteiro | Coeficiente Traseiro |
| :-------------- | :--: | :-------------------: | :------------------: |
| **BMW M4 GT3**  | 2021 |        7.9585         |        7.9585        |

### Challengers Pack – 2022 Rear engine

| Nome do Veículo                 | Ano  | Coeficiente Dianteiro | Coeficiente Traseiro |
| :------------------------------ | :--: | :-------------------: | :------------------: |
| **Audi R8 LMS Evo II**          | 2022 |        7.5980         |        7.4855        |
| **BMW M2 Cup**                  | 2020 |        7.2886         |       10.0000        |
| **Ferrari 488 Challenge Evo**   | 2020 |        7.5980         |        7.4855        |
| **Lamborghini Huracan ST Evo2** | 2021 |        7.5980         |        7.4855        |
| **Porsche 992 GT3 Cup**         | 2021 |        7.1497         |        6.7715        |

## Appendix 4 – brakeBias

### GT3 - 2018

| Nome do Veículo                    | Ano  | Dash Offset |
| :--------------------------------- | :--: | :---------: |
| **Aston Martin Vantage V12 GT3**   | 2013 |     -7      |
| **Audi R8 LMS**                    | 2015 |     -14     |
| **Bentley Continental GT3**        | 2015 |     -7      |
| **Bentley Continental GT3**        | 2018 |     -7      |
| **BMW M6 GT3**                     | 2017 |     -15     |
| **Emil Frey Jaguar G3**            | 2012 |     -7      |
| **Ferrari 488 GT3**                | 2018 |     -17     |
| **Honda NSX GT3**                  | 2017 |     -14     |
| **Lamborghini Gallardo G3 Reiter** | 2017 |     -14     |
| **Lamborghini Huracan GT3**        | 2015 |     -14     |
| **Lamborghini Huracan ST**         | 2015 |     -14     |
| **Lexus RCF GT3**                  | 2016 |     -14     |
| **McLaren 650S GT3**               | 2015 |     -17     |
| **Mercedes AMG GT3**               | 2015 |     -14     |
| **Nissan GTR Nismo GT3**           | 2015 |     -15     |
| **Nissan GTR Nismo GT3**           | 2018 |     -15     |
| **Porsche 991 GT3 R**              | 2018 |     -21     |
| **Porsche 991 II GT3 Cup**         | 2017 |     -5      |

### GT3 - 2019

| Nome do Veículo                 | Ano  | Dash Offset |
| :------------------------------ | :--: | :---------: |
| **Aston Martin V8 Vantage GT3** | 2019 |     -7      |
| **Audi R8 LMS Evo**             | 2019 |     -14     |
| **Honda NSX GT3 Evo**           | 2019 |     -14     |
| **Lamborghini Huracan GT3 EVO** | 2019 |     -14     |
| **McLaren 720S GT3**            | 2019 |     -17     |
| **Porsche 911 II GT3 R**        | 2019 |     -21     |

### GT4

| Nome do Veículo                  | Ano  | Dash Offset |
| :------------------------------- | :--: | :---------: |
| **Alpine A110 GT4**              | 2018 |     -15     |
| **Aston Martin Vantage AMR GT4** | 2018 |     -20     |
| **Audi R8 LMS GT4**              | 2016 |     -15     |
| **BMW M4 GT4**                   | 2018 |     -22     |
| **Chevrolet Camaro GT4 R**       | 2017 |     -18     |
| **Ginetta G55 GT4**              | 2012 |     -18     |
| **Ktm Xbow GT4**                 | 2016 |     -20     |
| **Maserati Gran Turismo MC GT4** | 2016 |     -15     |
| **McLaren 570s GT4**             | 2016 |     -9      |
| **Mercedes AMG GT4**             | 2016 |     -20     |
| **Porsche 718 Cayman GT4 MR**    | 2019 |     -20     |

### GT3 – 2020

| Nome do Veículo          | Ano  | Dash Offset |
| :----------------------- | :--: | :---------: |
| **Ferrari 488 GT3 Evo**  | 2020 |     -17     |
| **Mercedes AMG GT3 Evo** | 2020 |     -14     |

### GT3 – 2021

| Nome do Veículo | Ano  | Dash Offset |
| :-------------- | :--: | :---------: |
| **BMW M4 GT3**  | 2021 |     -14     |

### Challengers Pack – 2022

| Nome do Veículo                 | Ano  | Dash Offset |
| :------------------------------ | :--: | :---------: |
| **Audi R8 LMS Evo II**          | 2022 |     -14     |
| **BMW M2 Cup**                  | 2020 |     -17     |
| **Ferrari 488 Challenge Evo**   | 2020 |     -13     |
| **Lamborghini Huracan ST Evo2** | 2021 |     -14     |
| **Porsche 992 GT3 Cup**         | 2021 |     -5      |

## Appendix 5 – Max Steering Angle

### GT3 - 2018

| Nome do Veículo                    | Ano  | Dash Offset |
| :--------------------------------- | :--: | :---------: |
| **Aston Martin Vantage V12 GT3**   | 2013 |     -7      |
| **Audi R8 LMS**                    | 2015 |     -14     |
| **Bentley Continental GT3**        | 2015 |     -7      |
| **Bentley Continental GT3**        | 2018 |     -7      |
| **BMW M6 GT3**                     | 2017 |     -15     |
| **Emil Frey Jaguar G3**            | 2012 |     -7      |
| **Ferrari 488 GT3**                | 2018 |     -17     |
| **Honda NSX GT3**                  | 2017 |     -14     |
| **Lamborghini Gallardo G3 Reiter** | 2017 |     -14     |
| **Lamborghini Huracan GT3**        | 2015 |     -14     |
| **Lamborghini Huracan ST**         | 2015 |     -14     |
| **Lexus RCF GT3**                  | 2016 |     -14     |
| **McLaren 650S GT3**               | 2015 |     -17     |
| **Mercedes AMG GT3**               | 2015 |     -14     |
| **Nissan GTR Nismo GT3**           | 2015 |     -15     |
| **Nissan GTR Nismo GT3**           | 2018 |     -15     |
| **Porsche 991 GT3 R**              | 2018 |     -21     |
| **Porsche 991 II GT3 Cup**         | 2017 |     -5      |

### GT3 - 2019

| Nome do Veículo                 | Ano  | Ângulo (Graus) |
| :------------------------------ | :--: | :------------: |
| **Aston Martin V8 Vantage GT3** | 2019 |      320       |
| **Audi R8 LMS Evo**             | 2019 |      360       |
| **Honda NSX GT3 Evo**           | 2019 |      310       |
| **Lamborghini Huracan GT3 EVO** | 2019 |      310       |
| **McLaren 720S GT3**            | 2019 |      240       |
| **Porsche 911 II GT3 R**        | 2019 |      400       |

### GT4

| Nome do Veículo                  | Ano  | Ângulo (Graus) |
| :------------------------------- | :--: | :------------: |
| **Alpine A110 GT4**              | 2018 |      360       |
| **Aston Martin Vantage AMR GT4** | 2018 |      320       |
| **Audi R8 LMS GT4**              | 2016 |      360       |
| **BMW M4 GT4**                   | 2018 |      246       |
| **Chevrolet Camaro GT4 R**       | 2017 |      360       |
| **Ginetta G55 GT4**              | 2012 |      360       |
| **Ktm Xbow GT4**                 | 2016 |      290       |
| **Maserati Gran Turismo MC GT4** | 2016 |      450       |
| **McLaren 570s GT4**             | 2016 |      240       |
| **Mercedes AMG GT4**             | 2016 |      246       |
| **Porsche 718 Cayman GT4 MR**    | 2019 |      400       |

### GT3 – 2020

| Nome do Veículo          | Ano  | Ângulo (Graus) |
| :----------------------- | :--: | :------------: |
| **Ferrari 488 GT3 Evo**  | 2020 |      240       |
| **Mercedes AMG GT3 Evo** | 2020 |      320       |

### GT3 – 2021

| Nome do Veículo | Ano  | Ângulo (Graus) |
| :-------------- | :--: | :------------: |
| **BMW M4 GT3**  | 2021 |      270       |

### Challengers Pack – 2022

| Nome do Veículo                 | Ano  | Ângulo (Graus) |
| :------------------------------ | :--: | :------------: |
| **Audi R8 LMS Evo II**          | 2022 |      360       |
| **BMW M2 Cup**                  | 2020 |      180       |
| **Ferrari 488 Challenge Evo**   | 2020 |      240       |
| **Lamborghini Huracan ST Evo2** | 2021 |      310       |
| **Porsche 992 GT3 Cup**         | 2021 |      270       |

## Appendix 6 – CarModelId

### GT3 - 2018

| Nome do Veículo                    | Ano  | CarModelId |
| :--------------------------------- | :--: | :--------: |
| **Aston Martin Vantage V12 GT3**   | 2013 |     12     |
| **Audi R8 LMS**                    | 2015 |     3      |
| **Bentley Continental GT3**        | 2015 |     11     |
| **Bentley Continental GT3**        | 2018 |     8      |
| **BMW M6 GT3**                     | 2017 |     7      |
| **Emil Frey Jaguar G3**            | 2012 |     14     |
| **Ferrari 488 GT3**                | 2018 |     2      |
| **Honda NSX GT3**                  | 2017 |     17     |
| **Lamborghini Gallardo G3 Reiter** | 2017 |     13     |
| **Lamborghini Huracan GT3**        | 2015 |     4      |
| **Lamborghini Huracan ST**         | 2015 |     18     |
| **Lexus RCF GT3**                  | 2016 |     15     |
| **McLaren 650S GT3**               | 2015 |     5      |
| **Mercedes AMG GT3**               | 2015 |     1      |
| **Nissan GTR Nismo GT3**           | 2015 |     10     |
| **Nissan GTR Nismo GT3**           | 2018 |     6      |
| **Porsche 991 GT3 R**              | 2018 |     0      |
| **Porsche 991 II GT3 Cup**         | 2017 |     9      |

### GT3 - 2019

| Nome do Veículo                 | Ano  | CarModelId |
| :------------------------------ | :--: | :--------: |
| **Aston Martin V8 Vantage GT3** | 2019 |     20     |
| **Audi R8 LMS Evo**             | 2019 |     19     |
| **Honda NSX GT3 Evo**           | 2019 |     21     |
| **Lamborghini Huracan GT3 EVO** | 2019 |     16     |
| **McLaren 720S GT3**            | 2019 |     22     |
| **Porsche 911 II GT3 R**        | 2019 |     23     |

### GT4

| Nome do Veículo                  | Ano  | CarModelId |
| :------------------------------- | :--: | :--------: |
| **Alpine A110 GT4**              | 2018 |     50     |
| **Aston Martin Vantage AMR GT4** | 2018 |     51     |
| **Audi R8 LMS GT4**              | 2016 |     52     |
| **BMW M4 GT4**                   | 2018 |     53     |
| **Chevrolet Camaro GT4 R**       | 2017 |     55     |
| **Ginetta G55 GT4**              | 2012 |     56     |
| **Ktm Xbow GT4**                 | 2016 |     57     |
| **Maserati Gran Turismo MC GT4** | 2016 |     58     |
| **McLaren 570s GT4**             | 2016 |     59     |
| **Mercedes AMG GT4**             | 2016 |     60     |
| **Porsche 718 Cayman GT4 MR**    | 2019 |     61     |

### GT3 – 2020

| Nome do Veículo          | Ano  | CarModelId |
| :----------------------- | :--: | :--------: |
| **Ferrari 488 GT3 Evo**  | 2020 |     24     |
| **Mercedes AMG GT3 Evo** | 2020 |     25     |

### GT3 – 2021

| Nome do Veículo | Ano  | CarModelId |
| :-------------- | :--: | :--------: |
| **BMW M4 GT3**  | 2021 |     30     |

### Challengers Pack – 2022

| Nome do Veículo                 | Ano  | CarModelId |
| :------------------------------ | :--: | :--------: |
| **Audi R8 LMS Evo II**          | 2022 |     31     |
| **BMW M2 Cup**                  | 2020 |     27     |
| **Ferrari 488 Challenge Evo**   | 2020 |     26     |
| **Lamborghini Huracan ST Evo2** | 2021 |     29     |
| **Porsche 992 GT3 Cup**         | 2021 |     28     |

## Appendix 7 – Max RPM

### GT3 - 2018

| Nome do Veículo                    | Ano  | Max RPM |
| :--------------------------------- | :--: | :-----: |
| **Aston Martin Vantage V12 GT3**   | 2013 |  7750   |
| **Audi R8 LMS**                    | 2015 |  8650   |
| **Bentley Continental GT3**        | 2015 |  7500   |
| **Bentley Continental GT3**        | 2018 |  7400   |
| **BMW M6 GT3**                     | 2017 |  7100   |
| **Emil Frey Jaguar G3**            | 2012 |  8750   |
| **Ferrari 488 GT3**                | 2018 |  7300   |
| **Honda NSX GT3**                  | 2017 |  7500   |
| **Lamborghini Gallardo G3 Reiter** | 2017 |  8650   |
| **Lamborghini Huracan GT3**        | 2015 |  8650   |
| **Lamborghini Huracan ST**         | 2015 |  8650   |
| **Lexus RCF GT3**                  | 2016 |  7750   |
| **McLaren 650S GT3**               | 2015 |  7500   |
| **Mercedes AMG GT3**               | 2015 |  7900   |
| **Nissan GTR Nismo GT3**           | 2015 |  7500   |
| **Nissan GTR Nismo GT3**           | 2018 |  7500   |
| **Porsche 991 GT3 R**              | 2018 |  9250   |
| **Porsche 991 II GT3 Cup**         | 2017 |  8500   |

### GT3 - 2019

| Nome do Veículo                 | Ano  | Max RPM |
| :------------------------------ | :--: | :-----: |
| **Aston Martin V8 Vantage GT3** | 2019 |  7250   |
| **Audi R8 LMS Evo**             | 2019 |  8650   |
| **Honda NSX GT3 Evo**           | 2019 |  7650   |
| **Lamborghini Huracan GT3 EVO** | 2019 |  8650   |
| **McLaren 720S GT3**            | 2019 |  7700   |
| **Porsche 911 II GT3 R**        | 2019 |  9250   |

### GT4

| Nome do Veículo                  | Ano  | Max RPM |
| :------------------------------- | :--: | :-----: |
| **Alpine A110 GT4**              | 2018 |  6450   |
| **Aston Martin Vantage AMR GT4** | 2018 |  7000   |
| **Audi R8 LMS GT4**              | 2016 |  8650   |
| **BMW M4 GT4**                   | 2018 |  7600   |
| **Chevrolet Camaro GT4 R**       | 2017 |  7500   |
| **Ginetta G55 GT4**              | 2012 |  7200   |
| **Ktm Xbow GT4**                 | 2016 |  6500   |
| **Maserati Gran Turismo MC GT4** | 2016 |  7000   |
| **McLaren 570s GT4**             | 2016 |  7600   |
| **Mercedes AMG GT4**             | 2016 |  7000   |
| **Porsche 718 Cayman GT4 MR**    | 2019 |  7800   |

### GT3 – 2020

| Nome do Veículo          | Ano  | Max RPM |
| :----------------------- | :--: | :-----: |
| **Ferrari 488 GT3 Evo**  | 2020 |  7600   |
| **Mercedes AMG GT3 Evo** | 2020 |  7600   |

### GT3 – 2021

| Nome do Veículo | Ano  | Max RPM |
| :-------------- | :--: | :-----: |
| **BMW M4 GT3**  | 2021 |  7000   |

### Challengers Pack – 2022

| Nome do Veículo                 | Ano  | Max RPM |
| :------------------------------ | :--: | :-----: |
| **Audi R8 LMS Evo II**          | 2022 |  8650   |
| **BMW M2 Cup**                  | 2020 |  7520   |
| **Ferrari 488 Challenge Evo**   | 2020 |  8000   |
| **Lamborghini Huracan ST Evo2** | 2021 |  8650   |
| **Porsche 992 GT3 Cup**         | 2021 |  8750   |
