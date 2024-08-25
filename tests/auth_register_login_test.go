package tests

import (
	"sso/tests/suite"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt"
	ssov1 "github.com/jacute/protos/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID int32 = 0
	appID      int32 = 1
	appSecret        = "test-secret"

	passwordDefaultLen = 10
	expDeltaSeconds    = 5
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email, password := randomCredentials()
	resRegister, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	userID := resRegister.GetUserId()
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	resLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := resLogin.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, userID, int64(claims["userID"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int32(claims["appID"].(float64)))

	assert.InDelta(t, loginTime.Add(st.Config.TokenTTL).Unix(), claims["exp"].(float64), expDeltaSeconds)
}

func TestRegister_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	email, password := randomCredentials()
	resRegister, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	userID := resRegister.GetUserId()
	require.NoError(t, err)
	assert.NotEmpty(t, userID)

	resSecondRegister, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, resSecondRegister.GetUserId())
	assert.ErrorContains(t, err, "User already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	cases := []struct {
		name     string
		email    string
		password string
		want     string
	}{
		{
			name:     "Empty password",
			email:    gofakeit.Email(),
			password: "",
			want:     "Field 'Password' is required",
		},
		{
			name:     "Empty email",
			email:    "",
			password: gofakeit.Password(true, true, true, true, true, passwordDefaultLen),
			want:     "Field 'Email' is required",
		},
		{
			name:     "Empty email and password",
			email:    "",
			password: "",
			want:     "Field 'Email' is required",
		},
		{
			name:     "Short password",
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, true, 7),
			want:     "Field 'Password' require minimum 8 characters",
		},
		{
			name:     "Invalid Email",
			email:    gofakeit.Word(),
			password: gofakeit.Password(true, true, true, true, true, passwordDefaultLen),
			want:     "Field 'Email' is invalid",
		},
	}

	var wg sync.WaitGroup
	for _, c := range cases {
		wg.Add(1)
		go t.Run(c.name, func(t *testing.T) {
			defer wg.Done()
			res, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    c.email,
				Password: c.password,
			})

			require.Error(t, err)
			assert.Empty(t, res.GetUserId())
			assert.ErrorContains(t, err, c.want)
		})
	}
	wg.Wait()
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)
	cases := []struct {
		name     string
		email    string
		password string
		appID    int32
		want     string
	}{
		{
			name:     "Login with empty email",
			email:    "",
			password: gofakeit.Password(true, true, true, true, true, 10),
			appID:    appID,
			want:     "Field 'Email' is required",
		},
		{
			name:     "Login with empty password",
			email:    gofakeit.Email(),
			password: "",
			appID:    appID,
			want:     "Field 'Password' is required",
		},
		{
			name:     "Login with empty appID",
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, true, 10),
			appID:    emptyAppID,
			want:     "Field 'AppID' is required",
		},
		{
			name:     "Login with negative appID",
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, true, 10),
			appID:    -1,
			want:     "Field 'AppID' is invalid",
		},
		{
			name:     "Login in non-existent account",
			email:    gofakeit.Email(),
			password: gofakeit.Password(true, true, true, true, true, 10),
			appID:    appID,
			want:     "Invalid credentials",
		},
	}

	var wg sync.WaitGroup
	for _, c := range cases {
		wg.Add(1)
		go t.Run(c.name, func(t *testing.T) {
			defer wg.Done()

			res, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    c.email,
				Password: c.password,
				AppId:    c.appID,
			})
			require.Error(t, err)
			assert.Empty(t, res.GetToken())
			assert.ErrorContains(t, err, c.want)
		})
	}
	wg.Wait()
}

func randomCredentials() (string, string) {
	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, passwordDefaultLen)
	return email, password
}
