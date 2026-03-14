package entity

type Status struct {
	ID   int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Text string `gorm:"type:varchar(50);not null" json:"text"`

	Applications  []Application   `gorm:"foreignKey:StatusID" json:"applications,omitempty"`
	StatusHistory []StatusHistory `gorm:"foreignKey:StatusID" json:"status_history,omitempty"`
}
