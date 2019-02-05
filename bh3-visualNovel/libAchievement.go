package vn

// map 使用 key 1 和 2 区分 逆熵 和 幽兰戴尔
const (
	ANTIENTROPY = 1
	DURANDAL    = 2
)

type LIBAchievement struct {
	Lib map[int]VnAchievements
}

type VnAchievements struct {
	version  string
	Achieves map[string]achievementCode
}

type achievementCode struct {
	id      string
	chapter string
	scene   string
	action  string
	code    string
}

func (l *LIBAchievement) SetNovelAchievements(vNo int, vnA VnAchievements) {
	l.Lib[vNo] = vnA
}

func (l *LIBAchievement) GetNovelAchievements(vNo int) VnAchievements {
	return l.Lib[vNo]
}

func (l *LIBAchievement) getVersion(vNo int) string {
	return l.Lib[vNo].version
}

func (l *LIBAchievement) IsEmpty(vNo int) bool {
	t := l.Lib[vNo].Achieves
	if len(t) > 0 {
		return false
	}
	return true
}
