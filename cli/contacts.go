package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// AddContacts wires `fiken contacts {list,get,create,update,delete,persons}`.
//
// All subcommands are wired against the ogen Client. Mutating ones
// (create / update) read their JSON body from --from-file via the
// shared ReadBodyFile helper.
func AddContacts(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	_ = stderr
	set := ff.NewFlagSet("contacts")
	if parentFS, ok := root.Flags.(*ff.FlagSet); ok {
		set.SetParent(parentFS)
	}

	cmd := &ff.Command{
		Name:      "contacts",
		Usage:     "fiken contacts <subcommand>",
		ShortHelp: "Manage Fiken contacts.",
		Flags:     set,
	}

	// list
	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var (
		listPage     int
		listPageSize int
	)
	listSet.IntVar(&listPage, 0, "page", 0, "Page number (0-indexed)")
	listSet.IntVar(&listPageSize, 0, "page-size", 0, "Page size")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken contacts list --company <slug>",
		ShortHelp: "List contacts for a company.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ContactsList(ctx, ops.ContactsListIn{
				Company:  *sf.flagCompany,
				Page:     listPage,
				PageSize: listPageSize,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// get
	getSet := ff.NewFlagSet("get")
	getSet.SetParent(set)
	var getContactID int64
	getSet.Int64Var(&getContactID, 0, "contact-id", 0, "Contact id (required)")
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken contacts get --company <slug> --contact-id <id>",
		ShortHelp: "Get a single contact by id.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ContactsGet(ctx, ops.ContactsGetIn{
				Company:   *sf.flagCompany,
				ContactID: getContactID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// create
	createSet := ff.NewFlagSet("create")
	createSet.SetParent(set)
	var createFromFile string
	createSet.StringVar(&createFromFile, 0, "from-file", "", "Path to JSON body (required)")
	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "fiken contacts create --company <slug> --from-file <path>",
		ShortHelp: "Create a contact.",
		Flags:     createSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.Contact](createFromFile)
			if e != nil {
				e.Op = ops.OpContactsCreate
				return Renderer(ctx).Render(ops.Err[ops.ContactOut](e))
			}
			res := Client(ctx).ContactsCreate(ctx, ops.ContactsCreateIn{
				Company: *sf.flagCompany,
				Body:    body,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// update
	updateSet := ff.NewFlagSet("update")
	updateSet.SetParent(set)
	var (
		updateContactID int64
		updateFromFile  string
	)
	updateSet.Int64Var(&updateContactID, 0, "contact-id", 0, "Contact id (required)")
	updateSet.StringVar(&updateFromFile, 0, "from-file", "", "Path to JSON body (required)")
	updateCmd := &ff.Command{
		Name:      "update",
		Usage:     "fiken contacts update --company <slug> --contact-id <id> --from-file <path>",
		ShortHelp: "Update a contact.",
		Flags:     updateSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.Contact](updateFromFile)
			if e != nil {
				e.Op = ops.OpContactsUpdate
				return Renderer(ctx).Render(ops.Err[ops.ContactOut](e))
			}
			res := Client(ctx).ContactsUpdate(ctx, ops.ContactsUpdateIn{
				Company:   *sf.flagCompany,
				ContactID: updateContactID,
				Body:      body,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// delete
	deleteSet := ff.NewFlagSet("delete")
	deleteSet.SetParent(set)
	var deleteContactID int64
	deleteSet.Int64Var(&deleteContactID, 0, "contact-id", 0, "Contact id (required)")
	deleteCmd := &ff.Command{
		Name:      "delete",
		Usage:     "fiken contacts delete --company <slug> --contact-id <id>",
		ShortHelp: "Delete or deactivate a contact.",
		Flags:     deleteSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ContactsDelete(ctx, ops.ContactsDeleteIn{
				Company:   *sf.flagCompany,
				ContactID: deleteContactID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	personsCmd := addContactsPersons(set, stdout, stderr, sf)
	attachmentsCmd := contactsAttachmentsCmd(set, stdout, stderr, sf)

	cmd.Subcommands = []*ff.Command{listCmd, getCmd, createCmd, updateCmd, deleteCmd, personsCmd, attachmentsCmd}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}

// addContactsPersons wires `fiken contacts persons {list,get,create,update,delete}`
// as a nested subcommand. Mutating create / update read JSON bodies via
// --from-file.
func addContactsPersons(parent *ff.FlagSet, stdout, stderr io.Writer, sf *sessionFactory) *ff.Command {
	set := ff.NewFlagSet("persons")
	set.SetParent(parent)
	cmd := &ff.Command{
		Name:      "persons",
		Usage:     "fiken contacts persons <subcommand>",
		ShortHelp: "Manage contact persons.",
		Flags:     set,
	}

	// persons list
	listSet := ff.NewFlagSet("list")
	listSet.SetParent(set)
	var listContactID int64
	listSet.Int64Var(&listContactID, 0, "contact-id", 0, "Contact id (required)")
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken contacts persons list --company <slug> --contact-id <id>",
		ShortHelp: "List contact persons for a contact.",
		Flags:     listSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ContactsPersonsList(ctx, ops.ContactsPersonsListIn{
				Company:   *sf.flagCompany,
				ContactID: listContactID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// persons get
	getSet := ff.NewFlagSet("get")
	getSet.SetParent(set)
	var (
		getContactID       int64
		getContactPersonID int64
	)
	getSet.Int64Var(&getContactID, 0, "contact-id", 0, "Contact id (required)")
	getSet.Int64Var(&getContactPersonID, 0, "contact-person-id", 0, "Contact person id (required)")
	getCmd := &ff.Command{
		Name:      "get",
		Usage:     "fiken contacts persons get --company <slug> --contact-id <id> --contact-person-id <id>",
		ShortHelp: "Get a single contact person.",
		Flags:     getSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ContactsPersonsGet(ctx, ops.ContactsPersonsGetIn{
				Company:         *sf.flagCompany,
				ContactID:       getContactID,
				ContactPersonID: getContactPersonID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// persons create
	createSet := ff.NewFlagSet("create")
	createSet.SetParent(set)
	var (
		createContactID int64
		createFromFile  string
	)
	createSet.Int64Var(&createContactID, 0, "contact-id", 0, "Contact id (required)")
	createSet.StringVar(&createFromFile, 0, "from-file", "", "Path to JSON body (required)")
	createCmd := &ff.Command{
		Name:      "create",
		Usage:     "fiken contacts persons create --company <slug> --contact-id <id> --from-file <path>",
		ShortHelp: "Add a contact person.",
		Flags:     createSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.ContactPerson](createFromFile)
			if e != nil {
				e.Op = ops.OpContactsPersonsCreate
				return Renderer(ctx).Render(ops.Err[ops.ContactPersonOut](e))
			}
			res := Client(ctx).ContactsPersonsCreate(ctx, ops.ContactsPersonsCreateIn{
				Company:   *sf.flagCompany,
				ContactID: createContactID,
				Body:      body,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// persons update
	updateSet := ff.NewFlagSet("update")
	updateSet.SetParent(set)
	var (
		updateContactID       int64
		updateContactPersonID int64
		updateFromFile        string
	)
	updateSet.Int64Var(&updateContactID, 0, "contact-id", 0, "Contact id (required)")
	updateSet.Int64Var(&updateContactPersonID, 0, "contact-person-id", 0, "Contact person id (required)")
	updateSet.StringVar(&updateFromFile, 0, "from-file", "", "Path to JSON body (required)")
	updateCmd := &ff.Command{
		Name:      "update",
		Usage:     "fiken contacts persons update --company <slug> --contact-id <id> --contact-person-id <id> --from-file <path>",
		ShortHelp: "Update a contact person.",
		Flags:     updateSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			body, e := ReadBodyFile[fiken.ContactPerson](updateFromFile)
			if e != nil {
				e.Op = ops.OpContactsPersonsUpdate
				return Renderer(ctx).Render(ops.Err[ops.ContactPersonOut](e))
			}
			res := Client(ctx).ContactsPersonsUpdate(ctx, ops.ContactsPersonsUpdateIn{
				Company:         *sf.flagCompany,
				ContactID:       updateContactID,
				ContactPersonID: updateContactPersonID,
				Body:            body,
			})
			return Renderer(ctx).Render(res)
		},
	}

	// persons delete
	deleteSet := ff.NewFlagSet("delete")
	deleteSet.SetParent(set)
	var (
		deleteContactID       int64
		deleteContactPersonID int64
	)
	deleteSet.Int64Var(&deleteContactID, 0, "contact-id", 0, "Contact id (required)")
	deleteSet.Int64Var(&deleteContactPersonID, 0, "contact-person-id", 0, "Contact person id (required)")
	deleteCmd := &ff.Command{
		Name:      "delete",
		Usage:     "fiken contacts persons delete --company <slug> --contact-id <id> --contact-person-id <id>",
		ShortHelp: "Delete a contact person.",
		Flags:     deleteSet,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.Build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			res := Client(ctx).ContactsPersonsDelete(ctx, ops.ContactsPersonsDeleteIn{
				Company:         *sf.flagCompany,
				ContactID:       deleteContactID,
				ContactPersonID: deleteContactPersonID,
			})
			return Renderer(ctx).Render(res)
		},
	}

	cmd.Subcommands = []*ff.Command{listCmd, getCmd, createCmd, updateCmd, deleteCmd}
	return cmd
}
