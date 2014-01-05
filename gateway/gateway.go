package gateway

type Gateway struct {}

func New(addr string) (gw *Gateway) {
	return &Gateway{}
}

func(gw *Gateway) SetTLS(cert, key string) (err error) {
	return nil
}

func(gw *Gateway) Close() {
	
}

func(gw *Gateway) Serve() (err error) {
	return nil
}
