package assets

type Assets struct {
	Dora string
}

func NewAssets() Assets {
	return Assets{
		Dora: "../assets/dora",
	}
}
