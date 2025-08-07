package game

// createMockCardService creates a CardService with realistic cards for testing
func createMockCardService() *CardService {
	return &CardService{
		Leaders: []*Card{
			{ID: 1, Name: "The Usurper", Types: CardTypes{Subtype: "Leader"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 2, Name: "The Rightful Heir", Types: CardTypes{Subtype: "Leader"}, Base64Image: "data:image/jpeg;base64,test"},
		},
		Guardians: []*Card{
			{ID: 3, Name: "The Bodyguard", Types: CardTypes{Subtype: "Guardian"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 4, Name: "The Knight", Types: CardTypes{Subtype: "Guardian"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 5, Name: "The Protector", Types: CardTypes{Subtype: "Guardian"}, Base64Image: "data:image/jpeg;base64,test"},
		},
		Assassins: []*Card{
			{ID: 6, Name: "The Assassin", Types: CardTypes{Subtype: "Assassin"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 7, Name: "The Infiltrator", Types: CardTypes{Subtype: "Assassin"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 8, Name: "The Shadow", Types: CardTypes{Subtype: "Assassin"}, Base64Image: "data:image/jpeg;base64,test"},
		},
		Traitors: []*Card{
			{ID: 9, Name: "The Cultist", Types: CardTypes{Subtype: "Traitor"}, Base64Image: "data:image/jpeg;base64,test"},
			{ID: 10, Name: "The Spy", Types: CardTypes{Subtype: "Traitor"}, Base64Image: "data:image/jpeg;base64,test"},
		},
	}
}
