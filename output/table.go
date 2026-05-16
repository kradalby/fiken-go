package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

type tableRenderer struct {
	w io.Writer
	t ErrorTranslator
}

type tableRow interface {
	TableHeader() []string
	TableRow() []string
}

func (r *tableRenderer) Render(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	okField := rv.FieldByName("Ok")
	errField := rv.FieldByName("Error")
	if errField.IsValid() && !errField.IsNil() {
		return r.renderError(errField.Interface())
	}
	if !okField.IsValid() || okField.IsNil() {
		return fmt.Errorf("table renderer: empty envelope")
	}
	ok := okField.Elem().Interface()
	return r.renderOk(ok)
}

func (r *tableRenderer) renderError(errVal any) error {
	var code, message string
	ev := reflect.ValueOf(errVal).Elem()
	if f := ev.FieldByName("Code"); f.IsValid() {
		code = f.String()
	}
	if f := ev.FieldByName("Message"); f.IsValid() {
		message = f.String()
	}
	out := message
	if r.t != nil {
		out = r.t(code, message)
	}
	_, err := fmt.Fprintln(r.w, out)
	return err
}

func (r *tableRenderer) renderOk(ok any) error {
	rv := reflect.ValueOf(ok)
	itemsField := rv.FieldByName("Items")
	if itemsField.IsValid() {
		return r.renderItems(itemsField)
	}
	return r.renderSingle(ok)
}

func (r *tableRenderer) renderItems(items reflect.Value) error {
	if items.Len() == 0 {
		_, err := fmt.Fprintln(r.w, "(no results)")
		return err
	}
	first, ok := items.Index(0).Interface().(tableRow)
	if !ok {
		return fmt.Errorf("table renderer: %s does not implement TableRow/TableHeader",
			items.Index(0).Type())
	}
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, strings.Join(first.TableHeader(), "\t")); err != nil {
		return err
	}
	for i := 0; i < items.Len(); i++ {
		row, ok := items.Index(i).Interface().(tableRow)
		if !ok {
			return fmt.Errorf("table renderer: item %d wrong type", i)
		}
		if _, err := fmt.Fprintln(tw, strings.Join(row.TableRow(), "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (r *tableRenderer) renderSingle(ok any) error {
	tr, ok2 := ok.(tableRow)
	if !ok2 {
		return fmt.Errorf("table renderer: %T must implement TableRow/TableHeader", ok)
	}
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, strings.Join(tr.TableHeader(), "\t")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(tw, strings.Join(tr.TableRow(), "\t")); err != nil {
		return err
	}
	return tw.Flush()
}
