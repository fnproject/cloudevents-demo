package cloudevent

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TODO make (*CloudEvent).ToRequest / (*CloudEvent).ToResponse? for symmetry

// FromRequest handles an http transport binding cloud event, per
// https://github.com/cloudevents/spec/blob/v0.1/http-transport-binding.md#14-event-formats
// FromRequest currently handles a json event formatted message, as specified
// in the Content-Type header, falling back to binary format for all other
// formats. It may consume, but never closes req.Body. If the contentType is
// json and ce.Data is a typed object, it will be decoded into, otherwise it
// will be decoded into interface{} and must be type cast; extensions works
// similarly with map[string]interface{}, if unset will be decoded into
// map[string]string. If the contentType is an unhandled type, the data section
// will be of type string.
// TODO make note about extensions nuance b/w json & binary or resolve
func (ce *CloudEvent) FromRequest(req *http.Request) error {
	rct := req.Header.Get("Content-Type")
	ct, _, err := mime.ParseMediaType(rct)
	if err != nil {
		return err
	}

	// if this is a cloud events media type & an event format we understand,
	// then treat it as _structured_ mode; otherwise, fall back to binary.
	if ct == "application/cloudevents+json" {
		// decode the body to the entire cloud event object in structured mode
		err = json.NewDecoder(req.Body).Decode(&ce)
	} else {
		err = ce.handleBinaryData(ct, req.Body)
	}
	if err != nil {
		return err
	}

	// regardless of data format, headers may be cloud event headers. let's just
	// override anything we may have already [for structured format].
	return ce.handleHeaders(req.Header)
}

func (ce *CloudEvent) handleBinaryData(contentType string, body io.Reader) error {
	// TODO I guess we could handle other media types, too? currently only json is explicitly pita.
	//
	// in binary mode, the whole request body is the data section and
	// 'Content-Type' maps directly onto cloudevent.contentType -- if it's of
	// content type json in any way, we must decode the data section as json.
	ce.ContentType = contentType
	if contentType == "application/json" || strings.Contains(contentType, "+json") {
		err := json.NewDecoder(body).Decode(ce.Data)
		if err != nil {
			return err
		}
	} else {
		// TODO we could leave Data as an io.Reader, this is a needless copy and callers
		// can 100% decode the body directly themselves from an io.Reader fine

		var b bytes.Buffer // TODO allow this to be specified or provide a pool
		_, err := io.Copy(&b, body)
		if err != nil {
			return err
		}

		ce.Data = b.String()
	}

	return nil
}

func (ce *CloudEvent) handleHeaders(headers http.Header) error {
	for k, vs := range headers {
		// we have to json parse each of these, e.g. a string is a quoted string
		v := vs[0] // TODO(reed): no arrays, RIGHT?
		dec := json.NewDecoder(strings.NewReader(v))

		// go headers are weird, and will camel case things, it's easier if we
		// just put everything in lower case to match. e.g. CE-EventType will be
		// Ce-Eventtype if constructed using http.Header methods, but could also
		// be CE-EventType -- both are fine, really (nothing really matters?)
		// TODO extensions are still weird. cool.
		lk := strings.ToLower(k)
		var err error
		switch {
		case lk == "ce-eventtype":
			err = dec.Decode(&ce.EventType)
		case lk == "ce-eventtypeversion":
			err = dec.Decode(&ce.EventTypeVersion)
		case lk == "ce-cloudeventsversion":
			err = dec.Decode(&ce.CloudEventsVersion)
		case lk == "ce-source":
			err = dec.Decode(&ce.Source)
		case lk == "ce-eventid":
			err = dec.Decode(&ce.EventID)
		case lk == "ce-eventtime":
			err = dec.Decode(&ce.EventTime)
		case lk == "ce-schemaurl":
			err = dec.Decode(&ce.SchemaURL)
		case strings.HasPrefix(lk, "ce-x-"):
			// uncapitalize first letter, strip prefix, try to leave the rest as is (if go lets us? TODO)
			k = lowerFirst(k[5:])

			// TODO :ExtractFunc
		redo:
			switch ext := ce.Extensions.(type) {
			case nil:
				// NOTE: map[string]interface{} is the default behavior of json decoding, seems
				// wise to be consistent, but maybe we should do something else?
				ce.Extensions = make(map[string]interface{})
				goto redo // can't fallthrough in type switch, they're not as bad as the internet may tell you...
			case map[string]string:
				var s string
				if err = dec.Decode(&s); err == nil {
					ext[k] = s
				}
			case map[string]interface{}:
				var s interface{}
				if err = dec.Decode(&s); err == nil {
					ext[k] = s
				}
			default:
				// TODO uh, we could use reflect i guess. it's a hell of a lot easier to make them use json tho
				// TODO or we could just replace this with a map[string]string and not error
				err = errors.New("unsupported extensions type in binary http cloud event envelope, only map[string]string or map[string]interface{} supported in binary format")
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
