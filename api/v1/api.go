package api

type API struct{}

func NewAPI() *API {
	return &API{}
}

func (a *API) Join() error {
	return nil
}
