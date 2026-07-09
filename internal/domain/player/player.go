package player

type Attributes struct {
	Pace int
	Passing int
	Shooting int
	Altura int
	Peso int
	Impulso int
	Explosao int
	Fisico int
	FisicalStatus int
	Cabeceio int
	Cruzamento int
	Habilidade int
	Finalizacao int
	Dominio int
	Temperamento int
}

type Player struct {
	Id string
	Name string
	Age int
	Position string

	Attributes Attributes
}
