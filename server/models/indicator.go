package models

// Indicator is an individual result from an engine
type Indicator struct {
	ID             string `orm:"column(id);index;pk"`
	ProcessID      string `orm:"column(process_id);index"`
	ProcessEventID string `orm:"column(process_event_id);index"`
	Engine         string
	RuleName       string `orm:"column(rule_name)"`
	IndicatorType  string `orm:"column(indicator_type)"`
	Description    string
	ExtraInfo      string `orm:"column(extra_info)"`
	Score          int
}
