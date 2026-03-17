package models

//CarDamage
type CarDamage [5]float32

func (cd CarDamage) Front() float32 {
	return cd[0]
}

func (cd CarDamage) Rear() float32 {
	return cd[1]
}

func (cd CarDamage) Left() float32 {
	return cd[2]
}

func (cd CarDamage) Right() float32 {
	return cd[3]
}

func (cd CarDamage) Center() float32 {
	return cd[4]
}

// Vector
type Vector [3]float32

func (v Vector) X() float32 {
	return v[0]
}

func (v Vector) Y() float32 {
	return v[1]
}

func (v Vector) Z() float32 {
	return v[2]
}

//Wheels
type Wheels [4]float32

func (w Wheels) FrontLeft() float32 {
	return w[0]
}

func (w Wheels) FrontRight() float32 {
	return w[1]
}

func (w Wheels) RearLeft() float32 {
	return w[2]
}

func (w Wheels) RearRight() float32 {
	return w[3]
}

//TireContactPoints
type TireContactPoints [4]Vector

func (tcp TireContactPoints) FrontLeft() Vector {
	return tcp[0]
}

func (tcp TireContactPoints) FrontRight() Vector {
	return tcp[1]
}

func (tcp TireContactPoints) RearLeft() Vector {
	return tcp[2]
}

func (tcp TireContactPoints) RearRight() Vector {
	return tcp[3]
}