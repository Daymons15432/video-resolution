package encoder

type Profile string

const (
	Low  Profile = "low"
	Med  Profile = "med"
	High Profile = "high"
)

func ParseProfile(s string) Profile {
	switch s {
	case "low":
		return Low
	case "high":
		return High
	default:
		return Med
	}
}
