package importer

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type ImporterArg struct {
	AppName, PackageName, Imports string
	Shallow                       bool
}

// Importer is an interface implemented by all sysl importers
type Importer interface {
	// LoadFile reads in a file from path and returns the generated Sysl.
	LoadFile(path string) (string, error)
	// Load takes in a string in a format supported by an the importer
	// It returns the converted Sysl as a string.
	Load(content string) (string, error)
	// Configure allows the imported Sysl application name, package and import directories to be specified.
	Configure(arg *ImporterArg) (Importer, error)
}

// Formats lists all supported import formats
// TODO: Add all transform imports dynamically
var Formats = []Format{
	Grammar,
	OpenAPI3,
	OpenAPI2,
	XSD,
	Avro,
	SpannerSQL,
	SpannerSQLDir,
	Protobuf,
	ProtobufDir,
	Postgres,
	PostgresDir,
	MySQL,
	MySQLDir,
	BigQuery,
	JSONSchema,
}

// Factory takes in an absolute path and its contents (if path is a file) and returns an importer
// for the detected file type.
func Factory(path string, isDir bool, formatName string, content []byte, logger *logrus.Logger) (Importer, error) {
	var format Format
	if formatName != "" {
		for _, f := range Formats {
			if strings.EqualFold(formatName, f.Name) {
				format = f
				break
			}
		}
		if format.Name == "" {
			return nil, fmt.Errorf("an importer does not exist for %s", formatName)
		}
	} else {
		// TODO: Get rid of format autodetection
		ft, err := GuessFileType(path, isDir, content, Formats)
		if err != nil {
			return nil, err
		}
		format = ft
	}

	for _, f := range Formats {
		if f.Name != format.Name {
			continue
		}

		logger.Debugln("Detected " + f.Name)

		break
	}

	switch format.Name {
	case OpenAPI2.Name:
		logger.Debugln("Detected OpenAPI2")
		return MakeOpenAPI2Importer(logger, "", path), nil
	case OpenAPI3.Name:
		logger.Debugln("Detected OpenAPI3")
		// FIXME: revert back to the arrai importer once regression test is done.
		return NewOpenAPIV3Importer(logger), nil
	case XSD.Name:
		logger.Debugln("Detected XSD")
		return MakeXSDImporter(logger), nil
	case Grammar.Name:
		logger.Debugln("Detected grammar file")
		return nil, fmt.Errorf("importer disabled for: %s", format.Name)
	case Avro.Name:
		logger.Debugln("Detected Avro")
		return NewAvroImporter(logger), nil
	case SpannerSQL.Name, SpannerSQLDir.Name, Postgres.Name, PostgresDir.Name, MySQL.Name, MySQLDir.Name, BigQuery.Name:
		logger.Debugln("Detected SQL")
		return MakeSQLImporter(logger), nil
	case Protobuf.Name, ProtobufDir.Name:
		logger.Debugln("Detected Protobuf")
		return MakeProtobufImporter(logger), nil
	default:
		logger.Debugln("Defaulting to transform-based importing")
		return MakeTransformImporter(logger, format.Name), nil
	}
}
