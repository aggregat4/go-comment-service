# Website Comment Service

This is a go service that offers embeddable commenting functionality for
websites. One of the target uses is to be able to embed this into an otherwise
statically generated website like a blog.

You can configure a website that you want to add comments to. This service can
render those comments in an iframe under its posts.

The service allows users to submit comments and uses email based authentication
to validate the email address.

Comments need to be authenticated to be considered for approval as this service
is built to be operated in a GDPR-compliant fashion.

Users can see, export, edit and delete their comments.

Admins for a particular website can screen, approve or deny comments. Only
authenticated comments are considered.


## Embedding the Comments Page

To embed comments on your website, you'll need to:

1. Register your website with the comment service to get a service ID
2. Add an iframe element to your page where you want the comments to appear
3. Implement the iframe resizing listener for a seamless experience

### Basic Implementation

Add an iframe to your page using the following format:

```html
<iframe
  src="https://your-comment-service.com/servicces/{serviceId}/posts/{postKey}/comments"
  width="100%"
  style="border: none;"
  id="comments-iframe"
></iframe>
```

Replace:
- `your-comment-service.com` with your actual comment service domain
- `{serviceId}` with your registered service ID
- `{postKey}` with a unique identifier for the current page/post

### Automatic Height Adjustment

To ensure the iframe resizes automatically to fit its content without scrollbars, add this JavaScript to your page:

```javascript
window.addEventListener('message', function(e) {
    // Verify the message origin for security
    if (e.origin !== 'https://your-comment-service.com') return;
    // Check if it's a height update message
    if (e.data && e.data.type === 'comment-height') {
        const iframe = document.querySelector('.comment-frame');
        if (iframe) {
            iframe.style.height = e.data.height + 'px';
        }
    }
});
```

The comment service will automatically send height update messages whenever the content size changes, ensuring a seamless integration without iframe scrollbars.


## Privacy Laws, GDPR and this Project

It is impossible to satisfy privacy law requirements on a technical level alone.

This implementation aims to provide technical means that can enable a privacy
preserving and privacy protecting comment service, but the ultimate compliance
with one or more privacy laws is an organizational effort that is out of scope
of this project.

Privacy law is relevant for comments, because they are inherently personal data
as we capture the email address, optionally a name and website of the user and
the comment content itself and then publicly display that on various websites.

The user needs to have a set of tools to see, export, change and delete their
own personal data.

The user needs to be informed about the way that their data is used and shared
with third parties through a privacy policy.

An administrator needs the ability to screen and remove problematic content.

Finally, it must be possible to set age requirements on comment posting to avoid
running into consent issues when it comes to gathering personal data on minors.

## Security

### Encryption

All personal data (email addresses, optional name, optional website and comment
contents) is _encrypted at rest_.

The service should be operated over TLS through the use of an appropriate proxy
server.

### Authentication

Admins authenticate via OpenID Connect (OIDC) and require a service specific
admin claim.

Super admins authenticate via OIDC and require a "superadmin" claim to manage
the site configurations.

User authentication is based on email:

- Users are sent an email with a time-limited high entropy authentication token
- Upon token validation a time-limited cookie is set with the user's information
