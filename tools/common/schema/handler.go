// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package schema

import (
	"fmt"
	"slices"
	"strings"

	"github.com/urfave/cli"

	"go.temporal.io/server/common/log"
	dbschemas "go.temporal.io/server/schema"
)

// Setup sets up schema tables
func Setup(cli *cli.Context, db DB, logger log.Logger) error {
	cfg, err := newSetupConfig(cli, db)
	if err != nil {
		return err
	}
	return newSetupSchemaTask(db, cfg, logger).Run()
}

// Update updates the schema for the specified database
func Update(cli *cli.Context, db DB, logger log.Logger) error {
	cfg, err := newUpdateConfig(cli, db)
	if err != nil {
		return err
	}
	return newUpdateSchemaTask(db, cfg, logger).Run()
}

func newUpdateConfig(cli *cli.Context, db DB) (*UpdateConfig, error) {
	config := new(UpdateConfig)
	config.SchemaDir = cli.String(CLIOptSchemaDir)
	config.SchemaName = cli.String(CLIOptSchemaName)
	config.TargetVersion = cli.String(CLIOptTargetVersion)

	if err := validateUpdateConfig(config, db); err != nil {
		return nil, err
	}
	return config, nil
}

func newSetupConfig(cli *cli.Context, db DB) (*SetupConfig, error) {
	config := new(SetupConfig)
	config.SchemaFilePath = cli.String(CLIOptSchemaFile)
	config.SchemaName = cli.String(CLIOptSchemaName)
	config.InitialVersion = cli.String(CLIOptVersion)
	config.DisableVersioning = cli.Bool(CLIOptDisableVersioning)
	config.Overwrite = cli.Bool(CLIOptOverwrite)

	if err := validateSetupConfig(config, db); err != nil {
		return nil, err
	}
	return config, nil
}

func validateSetupConfig(config *SetupConfig, db DB) error {
	if len(config.SchemaFilePath) == 0 && len(config.SchemaName) == 0 && config.DisableVersioning {
		return NewConfigError("needs either " + flag(CLIOptSchemaFile) + " or " + flag(CLIOptSchemaName))
	}
	if (config.DisableVersioning && len(config.InitialVersion) > 0) ||
		(!config.DisableVersioning && len(config.InitialVersion) == 0) {
		return NewConfigError("missing argument; either " + flag(CLIOptDisableVersioning) + " or " +
			flag(CLIOptVersion) + " but not both must be specified")
	}
	if len(config.SchemaFilePath) > 0 && len(config.SchemaName) > 0 {
		return NewConfigError("either" + flag(CLIOptSchemaFile) + " or " +
			flag(CLIOptSchemaName) + " must be specified")
	}
	if len(config.SchemaName) > 0 {
		if !slices.Contains(dbschemas.PathsByDB(db.Type()), config.SchemaName) {
			return NewConfigError(fmt.Sprintf("%s must be one of: %v",
				flag(CLIOptSchemaName), dbschemas.PathsByDB(db.Type())))
		}
	}
	if !config.DisableVersioning {
		ver, err := normalizeVersionString(config.InitialVersion)
		if err != nil {
			return NewConfigError("invalid " + flag(CLIOptVersion) + " argument:" + err.Error())
		}
		config.InitialVersion = ver
	}
	return nil
}

func validateUpdateConfig(config *UpdateConfig, db DB) error {
	if len(config.SchemaDir) == 0 && len(config.SchemaName) == 0 {
		return NewConfigError("missing argument; either" + flag(CLIOptSchemaDir) + " or " +
			flag(CLIOptSchemaName) + " must be specified")
	}
	if len(config.SchemaDir) > 0 && len(config.SchemaName) > 0 {
		return NewConfigError("either" + flag(CLIOptSchemaDir) + " or " +
			flag(CLIOptSchemaName) + " must be specified")
	}
	if len(config.SchemaName) > 0 {
		if !slices.Contains(dbschemas.PathsByDB(db.Type()), config.SchemaName) {
			return NewConfigError(fmt.Sprintf("%s must be one of: %v",
				flag(CLIOptSchemaName), dbschemas.PathsByDB(db.Type())))
		}
	}
	if len(config.TargetVersion) > 0 {
		ver, err := normalizeVersionString(config.TargetVersion)
		if err != nil {
			return NewConfigError("invalid " + flag(CLIOptTargetVersion) + " argument:" + err.Error())
		}
		config.TargetVersion = ver
	}
	return nil
}

func flag(opt string) string {
	return "(-" + opt + ")"
}

func schemaFileEnding(schemaName string) string {
	if strings.Contains(schemaName, "cassandra") {
		return ".cql"
	}
	return ".sql"
}
