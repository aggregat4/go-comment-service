package server

import (
	"aggregat4/go-commentservice/internal/domain"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestStatus(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	res, err := http.Get(createServerUrl(serverConfig.Port, "/status"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/plain; charset=UTF-8", res.Header.Get("Content-Type"))
	body := readBody(res)
	assert.Equal(t, "OK", body)
}

func TestInvalidService(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	res, err := http.Get(createServerUrl(serverConfig.Port, "/services/foo/posts/bar/comments"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 404, res.StatusCode)
}

func TestEmptyCommentsPage(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	res, err := http.Get(createServerUrl(serverConfig.Port, "/services/"+TEST_SERVICE+"/posts/bar/comments"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
	//assert.Equal(t, "no-store", res.Header.Get("Cache-Control"))
	//assert.Contains(t, readBody(res), fmt.Sprintf("value=\"%s\"", TestState))
	body := readBody(res)
	assert.Contains(t, body, "<h1>Comments</h1>")
	assert.Contains(t, body, "<dl class=\"comments\">")
	assert.NotContains(t, body, "<dt>")
	assert.NotContains(t, body, "<dd>")
}

func TestSingleCommentPostPage(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	checkCommentExistenceForPost(t, TEST_POSTKEY1, TEST_COMMENT_PENDING_AUTHENTICATION, true)
}

func checkCommentExistenceForPost(t *testing.T, postKey string, comment string, shouldContain bool) {
	res, err := http.Get(createServerUrl(serverConfig.Port, "/services/"+TEST_SERVICE+"/posts/"+postKey+"/comments"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
	//assert.Equal(t, "no-store", res.Header.Get("Cache-Control"))
	//assert.Contains(t, readBody(res), fmt.Sprintf("value=\"%s\"", TestState))
	body := readBody(res)
	assert.Contains(t, body, "<h1>Comments</h1>")
	assert.Contains(t, body, "<dl class=\"comments\">")
	if shouldContain {
		assert.Contains(t, body, "<dd>"+comment, "Comment should be displayed")
	} else {
		assert.NotContains(t, body, "<dd>"+comment, "Comment should not be displayed")
	}
}

func TestUserAuthenticationForm(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	res, err := http.Get(createServerUrl(serverConfig.Port, "/userauthentication/"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
	body := readBody(res)
	assert.Contains(t, body, "<h1>Request Authentication Token</h1>")
	assert.Contains(t, body, "<form action=\"/userauthentication/\" method=\"POST\">")
}

func TestRequestAuthenticationLinkWithNoParams(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	res, err := http.Post(createServerUrl(serverConfig.Port, "/userauthentication"), "application/x-www-form-urlencoded", nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
}

func TestRequestAuthenticationLinkWithNonExistingEmailParam(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	formParams := url.Values{}
	formParams.Set("email", "foo@example.com")
	encodedParams := formParams.Encode()
	postBody := strings.NewReader(encodedParams)
	res, err := http.Post(
		createServerUrl(serverConfig.Port, "/userauthentication"),
		"application/x-www-form-urlencoded",
		postBody)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
	body := readBody(res)
	assert.Contains(t, body, "<h1>Request Authentication Token</h1>")
	assert.Contains(t, body, "<p class=\"error\">")
	assert.Equal(t, 0, controller.EmailSender.NumberOfEmailsSent, "EmailSender should NOT have been called")
}

func TestRequestAuthenticationLinkWithExistingEmailParam(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	formParams := url.Values{}
	formParams.Set("email", TEST_USER_NO_TOKEN)
	encodedParams := formParams.Encode()
	postBody := strings.NewReader(encodedParams)
	res, err := http.Post(
		createServerUrl(serverConfig.Port, "/userauthentication"),
		"application/x-www-form-urlencoded",
		postBody)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
	body := readBody(res)
	assert.Contains(t, body, "<h1>Request Authentication Token</h1>")
	assert.Contains(t, body, "<p class=\"success\">")
	assert.Contains(t, body, "An authentication token will be sent")
	assert.Equal(t, 1, controller.EmailSender.NumberOfEmailsSent, "EmailSender should have been called")
}

func TestUserAuthenticationWithUnknownToken(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	res, err := http.Get(createServerUrl(serverConfig.Port, "/userauthentication/INVALIDTOKEN"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
	body := readBody(res)
	assert.Contains(t, body, "<h1>Request Authentication Token</h1>")
	assert.Contains(t, body, "<p class=\"error\">")
	assert.Contains(t, body, "Invalid token")
}

func TestUserAuthenticationValidToken(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
}

func authenticateAndValidate(t *testing.T, client *http.Client, controller Controller, email string, token string) domain.User {
	res, err := client.Get(createServerUrl(serverConfig.Port, "/userauthentication/"+token))
	if err != nil {
		t.Fatal(err)
	}
	user, err := controller.Store.FindUserByEmail(email)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.Equal(t, "/users/"+strconv.Itoa(user.Id)+"/comments", res.Header.Get("Location"))
	return user
}

func TestUserAuthenticationExpiredToken(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	res, err := client.Get(createServerUrl(serverConfig.Port, "/userauthentication/"+TEST_AUTHTOKEN_EXPIRED))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.Equal(t, "/userauthentication/?error=Invalid+token", res.Header.Get("Location"))
}

func TestGetCommentForm(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	res, err := http.Get(createServerUrl(serverConfig.Port, "/services/"+TEST_SERVICE+"/posts/"+TEST_POSTKEY1+"/commentform"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
	body := readBody(res)
	assert.Contains(t, body, "<title>New Comment</title>")
	assert.Contains(t, body, "<h1>New Comment</h1>")
	assert.Contains(t, body, "<form method=\"POST\" action=\"/services/"+TEST_SERVICE+"/posts/"+TEST_POSTKEY1+"/comments\">")
}

func TestGetCommentFormWithExistingComment(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res, err := client.Get(createServerUrl(serverConfig.Port, "/services/"+TEST_SERVICE+"/posts/"+TEST_POSTKEY1+"/commentform?commentId="+strconv.Itoa(expectedCommentId)))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "text/html; charset=UTF-8", res.Header.Get("Content-Type"))
	body := readBody(res)
	assert.Contains(t, body, "<title>Edit Comment</title>")
	assert.Contains(t, body, "<h1>Edit Comment</h1>")
	assert.Contains(t, body, "<form method=\"POST\" action=\"/services/"+TEST_SERVICE+"/posts/"+TEST_POSTKEY1+"/comments\">")
	assert.Contains(t, body, "<input type=\"hidden\" name=\"commentId\" value=\""+strconv.Itoa(expectedCommentId)+"\">")
}

func postComment(t *testing.T, client *http.Client, formParams url.Values, postKey string) *http.Response {
	encodedParams := formParams.Encode()
	postBody := strings.NewReader(encodedParams)
	res, err := client.Post(
		createServerUrl(serverConfig.Port, "/services/"+TEST_SERVICE+"/posts/"+postKey+"/comments"),
		"application/x-www-form-urlencoded",
		postBody)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestCreateNewCommentUnauthenticated(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	formParams := url.Values{}
	formParams.Set("emailAddress", "foo@example.com")
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	comment := "This is a comment"
	formParams.Set("comment", comment)
	res := postComment(t, client, formParams, TEST_POSTKEY2)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.True(t, strings.HasPrefix(res.Header.Get("Location"), "/services/"+TEST_SERVICE+"/posts/"+TEST_POSTKEY2+"/comments"))
	checkCommentExistenceForPost(t, TEST_POSTKEY2, comment, true)
}

func TestCreateNewCommentAuthenticated(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	formParams := url.Values{}
	formParams.Set("emailAddress", TEST_USER_AUTHTOKEN_VALID)
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	comment := "This is a comment"
	formParams.Set("comment", comment)
	res := postComment(t, client, formParams, TEST_POSTKEY2)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.True(t, strings.HasPrefix(res.Header.Get("Location"), "/services/"+TEST_SERVICE+"/posts/"+TEST_POSTKEY2+"/comments"))
	checkCommentExistenceForPost(t, TEST_POSTKEY2, comment, true)
}

func TestCreateNewCommentWithMissingEmail(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	formParams := url.Values{}
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	comment := "This is a comment"
	formParams.Set("comment", comment)
	res := postComment(t, client, formParams, TEST_POSTKEY2)
	assert.Equal(t, 400, res.StatusCode)
}

func TestCreateNewCommentWithMissingComment(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	formParams := url.Values{}
	formParams.Set("emailAddress", "foo@example.com")
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	res := postComment(t, client, formParams, TEST_POSTKEY2)
	assert.Equal(t, 400, res.StatusCode)
}

func TestUpdateExistingCommentWithInvalidCommentId(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	formParams := url.Values{}
	formParams.Set("emailAddress", TEST_USER_AUTHTOKEN_VALID)
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	formParams.Set("commentId", "INVALIDCOMMENTID")
	comment := "This is a comment"
	formParams.Set("comment", comment)
	res := postComment(t, client, formParams, TEST_POSTKEY2)
	assert.Equal(t, 404, res.StatusCode)
}

func TestUpdateExistingCommentWithValidCommentId(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	formParams := url.Values{}
	formParams.Set("emailAddress", TEST_USER_AUTHTOKEN_VALID)
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	formParams.Set("commentId", strconv.Itoa(expectedCommentId))
	comment := "New Comment Contents"
	formParams.Set("comment", comment)
	res := postComment(t, client, formParams, TEST_POSTKEY1)
	assert.Equal(t, http.StatusFound, res.StatusCode)
	assert.True(t, strings.HasPrefix(res.Header.Get("Location"), "/services/"+TEST_SERVICE+"/posts/"+TEST_POSTKEY1+"/comments"))
	checkCommentExistenceForPost(t, TEST_POSTKEY1, comment, true)
}

func TestUpdateExistingCommentWithValidCommentIdButWrongUser(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID2, TEST_AUTHTOKEN_VALID2)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	formParams := url.Values{}
	formParams.Set("emailAddress", TEST_USER_AUTHTOKEN_VALID)
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	formParams.Set("commentId", strconv.Itoa(expectedCommentId))
	comment := "New Comment Contents"
	formParams.Set("comment", comment)
	res := postComment(t, client, formParams, TEST_POSTKEY1)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestUpdateExistingCommentWithoutAuthentication(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	formParams := url.Values{}
	formParams.Set("emailAddress", TEST_USER_AUTHTOKEN_VALID)
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	formParams.Set("commentId", strconv.Itoa(expectedCommentId))
	comment := "Modified Comment"
	formParams.Set("comment", comment)
	res := postComment(t, client, formParams, TEST_POSTKEY2)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestUpdateExistingCommentAsTheWrongUser(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID2, TEST_AUTHTOKEN_VALID2)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	formParams := url.Values{}
	formParams.Set("emailAddress", TEST_USER_AUTHTOKEN_VALID2)
	formParams.Set("name", "John Foo")
	formParams.Set("website", "http://example.com")
	formParams.Set("commentId", strconv.Itoa(expectedCommentId))
	comment := "Modified Comment"
	formParams.Set("comment", comment)
	res := postComment(t, client, formParams, TEST_POSTKEY1)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestDeleteExistingCommentWithValidCommentId(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	user := authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	checkCommentExistenceForPost(t, TEST_POSTKEY1, TEST_COMMENT_PENDING_AUTHENTICATION, true)
	res := deleteComment(t, client, user.Id, expectedCommentId)
	assert.Equal(t, http.StatusFound, res.StatusCode, "Deleting comment should redirect to the user's comment overview page")
	assert.True(t, strings.HasPrefix(res.Header.Get("Location"), "/users/"+strconv.Itoa(user.Id)+"/comments"), "Deleting comment should redirect to the user's comment overview page")
	checkCommentExistenceForPost(t, TEST_POSTKEY1, TEST_COMMENT_PENDING_AUTHENTICATION, false)
}

func deleteComment(t *testing.T, client *http.Client, userId int, commentId int) *http.Response {
	res, err := client.Post(
		createServerUrl(serverConfig.Port, "/users/"+strconv.Itoa(userId)+"/comments/"+strconv.Itoa(commentId)+"/delete"),
		"application/x-www-form-urlencoded",
		nil)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestDeleteExistingCommentWithInvalidCommentId(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	user := authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res := deleteComment(t, client, user.Id, expectedCommentId+1)
	// TODO update test with toast check when we implement it
	assert.Equal(t, http.StatusFound, res.StatusCode, "Deleting comment should redirect to the user's comment overview page")
}

func TestDeleteExistingCommentWithInvalidUserId(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	user := authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res := deleteComment(t, client, user.Id+1, expectedCommentId) // invalid user id
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Deleting another user's comments should fail with unauthorized")
}

func TestDeleteExistingCommentWithValidCommentIdButWrongUser(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient() // use a different client than the one used to create the comment
	user := authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID2, TEST_AUTHTOKEN_VALID2)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res := deleteComment(t, client, user.Id, expectedCommentId)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Deleting another user's comments should fail with unauthorized")
}

func TestDeleteExistingCommentWithoutAuthentication(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res := deleteComment(t, client, 1, expectedCommentId)
	assert.Equal(t, http.StatusFound, res.StatusCode, "Deleting comment without authentication should redirect to the userauthentication page")
	assert.True(t, strings.HasPrefix(res.Header.Get("Location"), "/userauthentication/"), "Deleting comment without authentication should redirect to the userauthentication page")
}

func confirmComment(t *testing.T, client *http.Client, userId int, commentId int) *http.Response {
	res, err := client.Post(
		createServerUrl(serverConfig.Port, "/users/"+strconv.Itoa(userId)+"/comments/"+strconv.Itoa(commentId)+"/confirm"),
		"application/x-www-form-urlencoded",
		nil)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestConfirmExistingCommentWithAuthentication(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	user := authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	comment, err := controller.Store.GetComment(expectedCommentId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, domain.CommentStatusPendingAuthentication, comment.Status, "Comment should be pending authentication")
	res := confirmComment(t, client, user.Id, expectedCommentId)
	assert.Equal(t, http.StatusFound, res.StatusCode, "Confirming comment should redirect to the user's comment overview page")
	assert.True(t, strings.HasPrefix(res.Header.Get("Location"), "/users/"), "Confirming comment should redirect to the user's comment overview page")
	comment, err = controller.Store.GetComment(expectedCommentId)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, domain.CommentStatusPendingApproval, comment.Status, "Confirming comment should change the status to pending approval")
}

func TestConfirmExistingCommentWithoutAuthentication(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res := confirmComment(t, client, 1, expectedCommentId)
	assert.Equal(t, http.StatusFound, res.StatusCode, "Confirming comment without authentication should redirect to the userauthentication page")
	assert.True(t, strings.HasPrefix(res.Header.Get("Location"), "/userauthentication/"), "Confirming comment without authentication should redirect to the userauthentication page")
}

func TestConfirmExistingCommentWithInvalidCommentId(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	user := authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res := confirmComment(t, client, user.Id, expectedCommentId+1)
	// TODO update test with toast check when we implement it to see whether we can confirm a comment that does not exist
	assert.Equal(t, http.StatusFound, res.StatusCode, "Confirming comment with invalid comment id should redirect to the user's comment overview page")
	assert.True(t, strings.HasPrefix(res.Header.Get("Location"), "/users/"), "Confirming comment with invalid comment id should redirect to the user's comment overview page")
}

func TestConfirmExistingCommentWithInvalidUserId(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	user := authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID, TEST_AUTHTOKEN_VALID)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res := confirmComment(t, client, user.Id+1, expectedCommentId)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "You can not confirm comments that are not yours")
}

func TestConfirmExistingCommentWithValidCommentIdButWrongUser(t *testing.T) {
	echoServer, controller := waitForServer(t)
	defer echoServer.Close()
	defer controller.Store.Close()
	client := createTestHttpClient()
	user := authenticateAndValidate(t, client, controller, TEST_USER_AUTHTOKEN_VALID2, TEST_AUTHTOKEN_VALID2)
	expectedCommentId := findCommentByContent(TEST_COMMENTS, TEST_COMMENT_PENDING_AUTHENTICATION).Id
	res := confirmComment(t, client, user.Id, expectedCommentId)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Confirming comment with invalid user id should redirect to the user's comment overview page")
}
