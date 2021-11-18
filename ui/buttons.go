package ui

type Card struct {
	Sections []*Section `json:"sections,omitempty"`
}

type Section struct {
	// Header: The header of the section, text formatted supported.
	Header string `json:"header,omitempty"`

	// Widgets: A section must contain at least 1 widget.
	Widgets []*Widget `json:"widgets,omitempty"`
}

type Widget struct {
	Text string `json:"text,omitempty"`
	// Buttons: A list of buttons. Buttons is also oneof data and only one
	// of these
	// fields should be set.
	Buttons []*Button `json:"buttons,omitempty"`

	// Image: Display an image in this widget.
	Image *Image `json:"image,omitempty"`

	// KeyValue: Display a key value item in this widget.
	// KeyValue *KeyValue `json:"keyValue,omitempty"`

	// TextParagraph: Display a text paragraph in this widget.
	// TextParagraph *TextParagraph `json:"textParagraph,omitempty"`
}
type Image struct {

	// URL The URL of the image.
	URL string `json:"url,omitempty"`

	// OnClick: The onclick action.
	// OnClick *OnClick `json:"onClick,omitempty"`
}

type Button struct {
	Text string `json:"text,omitempty"`
	Link string `json:"link,omitempty"`
	Data string `json:"callback_data,omitempty"` // Believe this is telegram specific

	// OnClick: The onclick action.
	// OnClick *OnClick `json:"onClick,omitempty"`
}
