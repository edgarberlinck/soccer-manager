package player

type Attributes struct {
	Pace int
	Passing int
	Shooting int
}

type Player struct {
	Id string
	Name string
	Age int

	Attributes Attributes
}
