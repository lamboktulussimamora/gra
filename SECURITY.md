# Security Policy

## Supported Versions

Currently, the following versions of GRA Framework are being supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0.0 | :x:                |

## Reporting a Vulnerability

The GRA Framework team takes security seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report a Vulnerability

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to [security@example.com](mailto:security@example.com) (replace with your actual security contact).

Please include:

1. The version of GRA Framework you're using
2. A detailed description of the vulnerability
3. Steps to reproduce the issue
4. Potential impact of the vulnerability
5. Any potential solutions you've identified

### What to Expect

- We will acknowledge receipt of your vulnerability report within 48 hours.
- We will provide a more detailed response within 7 days, including our assessment of the issue's validity.
- We will work with you to understand and resolve the issue, keeping you informed of our progress.
- After addressing the vulnerability, we will publicly disclose it in our release notes and credit you (unless you prefer to remain anonymous).

## Security Best Practices for GRA Framework

When using GRA Framework, consider the following security best practices:

1. **Keep Dependencies Updated**: Regularly update GRA Framework and its dependencies to receive security patches.

2. **Use Secure JWT Settings**:
   - Use strong secret keys for JWT tokens
   - Set appropriate token expiration times
   - Implement token refresh mechanisms
   - Store JWT secrets securely using environment variables

3. **Implement Rate Limiting**: Use rate limiting middleware to prevent abuse.

4. **Enable HTTPS**: Always use HTTPS in production environments.

5. **Input Validation**: Use the validator package to validate all user inputs.

6. **Error Handling**: Use proper error handling to avoid leaking sensitive information.

7. **Logging**: Be careful not to log sensitive information.

## Security Features

GRA Framework provides several security-focused features:

- Secure headers middleware
- CORS protection
- JWT authentication
- Request validation
- XSS protection
- Content Security Policy support
