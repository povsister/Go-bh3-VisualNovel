package vn


type Portrait struct {
	Name string
	Index int
}

type Achievement struct {
	Achievement []map[string]string
	Msg string
	Portrait []Portrait
	Progress interface{}
	Retcode float64
}

type AchievementSubmitted struct {
	Achievement string
	Msg string
	Retcode float64
}
