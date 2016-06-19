package core

func NewApplication() *Application {
	return &Application{
		Streams: make(map[string]*Stream),
	}
}

func (app *Application) AcquireStream(name string) *Stream {
	stream, ok := app.Streams[name]
	if !ok {
		stream = &Stream{
			Name: name,
		}
		app.Streams[name] = stream
	}
	return stream
}