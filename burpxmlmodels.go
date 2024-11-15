package burptomap

import (
	"encoding/base64"
	"encoding/xml"
)

type Items struct {
	XMLName     xml.Name `xml:"items"`
	BurpVersion string   `xml:"burpVersion,attr"`
	ExportTime  string   `xml:"exportTime,attr"`
	Items       []Item   `xml:"item"`
}

type Item struct {
	Time           string `xml:"time"`
	URL            string `xml:"url"`
	Host           Host   `xml:"host"`
	Port           int    `xml:"port"`
	Protocol       string `xml:"protocol"`
	Method         string `xml:"method"`
	Path           string `xml:"path"`
	Extension      string `xml:"extension"`
	Request        Data   `xml:"request"`
	Status         int    `xml:"status"`
	ResponseLength int    `xml:"responselength"`
	MimeType       string `xml:"mimetype"`
	Response       Data   `xml:"response"`
	Comment        string `xml:"comment"`
}

type Host struct {
	IP   string `xml:"ip,attr"`
	Name string `xml:",chardata"`
}

type Data struct {
	Base64 string `xml:"base64,attr"`
	Value  string `xml:",chardata"`
}

// UnmarshalXML is unmarshaling the struct and decodes base64 if Value is base64 encoded
func (d *Data) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	type Alias Data // Create an alias to avoid recursive calls to UnmarshalXML
	aux := &struct {
		Content string `xml:",chardata"`
		Base64  string `xml:"base64,attr"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := dec.DecodeElement(aux, &start); err != nil {
		return err
	}

	// Check if Base64 attribute is "true" and decode if necessary
	if aux.Base64 == "true" {
		decodedContent, err := base64.StdEncoding.DecodeString(aux.Content)
		if err != nil {
			return err
		}
		d.Value = string(decodedContent)
		aux.Base64 = "false"
	} else {
		d.Value = aux.Content
	}
	d.Base64 = aux.Base64

	return nil
}
