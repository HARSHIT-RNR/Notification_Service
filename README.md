ERP_Notification_Service
ERP Backend

Event 1: Signup Verification (Step 3)
-> Consumed Topic: notification.send-signup-verification
-> Action: Send a verification email to the provided admin_mail with a verification_token.

Event 2: Provisioning Started (Step 7)
-> Consumed Topic: notification.provisioning-started
-> Action: Send an informational email to the admin, letting them know the tenant provisioning process has begun.

Event 3: Password Setup (Step 11)
-> Consumed Topic: notification.send-password-setup (produced by the Authentication Service).
-> Action: Send an email with a secure, single-use link for the admin to set their initial password.

Event 4: Welcome Notification (Step 13)
-> Consumed Topic: notification.send-welcome
-> Action: Send a final "Welcome" email after the tenant is fully provisioned and active, likely containing a link to the login page.

