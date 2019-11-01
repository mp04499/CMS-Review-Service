package handler

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	"google.golang.org/api/option"

	firebase "firebase.google.com/go"
	"github.com/subosito/gotenv"
)

type credentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderx509CertURL string `json:"auth_provider_x509_cert_url"`
	Clientx509CertURL       string `json:"client_x509_cert_url"`
}

type review struct {
	UID      string  `json:"uid"`
	Title    string  `json:"title"`
	Subtitle string  `json:"subtitle"`
	Review   string  `json:"review"`
	Score    float64 `json:"score"`
}

func init() {
	gotenv.Load()
}

var rev review
var cred credentials

// Handler Serverless Exported Function
func Handler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	cred, err := getCredentials()
	opt := option.WithCredentialsJSON(cred)
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := app.Firestore(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&rev)

	if len(rev.UID) < 1 {
		http.Error(w, "No UID Provided", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Body Error"+err.Error(), http.StatusBadRequest)
		return
	}

	score := math.Round(rev.Score*10) / 10
	db.Collection("users").Doc(rev.UID).Collection("reviews").Add(ctx, map[string]interface{}{
		"title":    rev.Title,
		"subtitle": rev.Subtitle,
		"review":   rev.Review,
		"score":    score,
	})
	fmt.Fprintf(w, "Review Creation Complete")

}

func getCredentials() ([]byte, error) {

	sa, err := b64.StdEncoding.DecodeString(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err == nil {
		json.Unmarshal(sa, &cred)
		credStruct := credentials(cred)
		credMarshal, err := json.Marshal(credStruct)

		return credMarshal, err
	}
	return nil, err
}
