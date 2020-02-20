package books

import (
	"bytes"
	"encoding/xml"
	"time"
)

type BookForm struct {
	Author  string    `xml:"urn:books author"`
	Title   string    `xml:"urn:books title"`
	Genre   string    `xml:"urn:books genre"`
	Price   float32   `xml:"urn:books price"`
	Pubdate time.Time `xml:"urn:books pub_date"`
	Review  string    `xml:"urn:books review"`
	Name    string    `xml:"urn:books name,attr,omitempty"`
}

func (t *BookForm) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type T BookForm
	var layout struct {
		*T
		Pubdate *xsdDate `xml:"urn:books pub_date"`
	}
	layout.T = (*T)(t)
	layout.Pubdate = (*xsdDate)(&layout.T.Pubdate)
	return e.EncodeElement(layout, start)
}
func (t *BookForm) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type T BookForm
	var overlay struct {
		*T
		Pubdate *xsdDate `xml:"urn:books pub_date"`
	}
	overlay.T = (*T)(t)
	overlay.Pubdate = (*xsdDate)(&overlay.T.Pubdate)
	return d.DecodeElement(&overlay, &start)
}

type BooksForm struct {
	Book []BookForm `xml:"urn:books book,omitempty"`
}

type xsdDate time.Time

func (t *xsdDate) UnmarshalText(text []byte) error {
	return _unmarshalTime(text, (*time.Time)(t), "2006-01-02")
}
func (t xsdDate) MarshalText() ([]byte, error) {
	return []byte((time.Time)(t).Format("2006-01-02")), nil
}
func (t xsdDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if (time.Time)(t).IsZero() {
		return nil
	}
	m, err := t.MarshalText()
	if err != nil {
		return err
	}
	return e.EncodeElement(m, start)
}
func (t xsdDate) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if (time.Time)(t).IsZero() {
		return xml.Attr{}, nil
	}
	m, err := t.MarshalText()
	return xml.Attr{Name: name, Value: string(m)}, err
}
func _unmarshalTime(text []byte, t *time.Time, format string) (err error) {
	s := string(bytes.TrimSpace(text))
	*t, err = time.Parse(format, s)
	if _, ok := err.(*time.ParseError); ok {
		*t, err = time.Parse(format+"Z07:00", s)
	}
	return err
}