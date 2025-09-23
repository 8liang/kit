package timetunnel

type Period string

const (
	Monthly  Period = "monthly"
	Weekly   Period = "weekly"
	Daily    Period = "daily"
	Hourly   Period = "hourly"
	Minutely Period = "minutely"
)
