package cmd

import (
	"database/sql"

	// import the crdbpgx for automatic retries of errors for crdb that support retry
	_ "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.hollow.sh/toolbox/ginjwt"

	"go.hollow.sh/serverservice/internal/dbtools"
	"go.hollow.sh/serverservice/internal/httpsrv"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts the hollow server",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().String("listen", "0.0.0.0:8000", "address to listen on")
	viperBindFlag("listen", serveCmd.Flags().Lookup("listen"))

	serveCmd.Flags().Bool("oidc", true, "use oidc auth")
	viperBindFlag("oidc.enabled", serveCmd.Flags().Lookup("oidc"))
	serveCmd.Flags().String("oidc-aud", "", "expected audience on OIDC JWT")
	viperBindFlag("oidc.audience", serveCmd.Flags().Lookup("oidc-aud"))
	serveCmd.Flags().String("oidc-issuer", "", "expected issuer of OIDC JWT")
	viperBindFlag("oidc.issuer", serveCmd.Flags().Lookup("oidc-issuer"))
	serveCmd.Flags().String("oidc-jwksuri", "", "URI for JWKS listing for JWTs")
	viperBindFlag("oidc.jwksuri", serveCmd.Flags().Lookup("oidc-jwksuri"))
	serveCmd.Flags().String("oidc-roles-claim", "claim", "field containing the permissions of an OIDC JWT")
	viperBindFlag("oidc.claims.roles", serveCmd.Flags().Lookup("oidc-roles-claim"))
	serveCmd.Flags().String("oidc-username-claim", "", "additional fields to output in logs from the JWT token, ex (email)")
	viperBindFlag("oidc.claims.username", serveCmd.Flags().Lookup("oidc-username-claim"))
}

func serve() {
	db, err := sql.Open("postgres", viper.GetString("db.uri"))
	if err != nil {
		logger.Fatalw("failed to init data store", "error", err)
	}

	if _, err := db.Exec("select 1;"); err != nil {
		logger.Fatalw("failed verifying database connection", "error", err)
	}

	dbtools.RegisterHooks()

	logger.Infow("starting server",
		"address", viper.GetString("listen"),
	)

	hs := &httpsrv.Server{
		Logger: logger.Desugar(),
		Listen: viper.GetString("listen"),
		Debug:  viper.GetBool("logging.debug"),
		DB:     db,
		AuthConfig: ginjwt.AuthConfig{
			Enabled:       viper.GetBool("oidc.enabled"),
			Audience:      viper.GetString("oidc.audience"),
			Issuer:        viper.GetString("oidc.issuer"),
			JWKSURI:       viper.GetString("oidc.jwksuri"),
			LogFields:     viper.GetStringSlice("oidc.log"),
			RolesClaim:    viper.GetString("oidc.claims.roles"),
			UsernameClaim: viper.GetString("oidc.claims.username"),
		},
	}

	if err := hs.Run(); err != nil {
		logger.Fatalw("failed starting server", "error", err)
	}
}
