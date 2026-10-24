package core

func AppPreferences(p Preferences) Preferences {
	p.Notifications = "all"
	p.RadiusMeters = ProximityBands()[0]
	p.BusinessNotifications = true
	return p
}
