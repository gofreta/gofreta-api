package emails

const ResetPasswordBody = `
{{define "content"}}
    <p>Hello,</p>

    <p>We've received a request to reset your account password.</p>

    <p>Your reset password token is:</p>

    <p class="text-center emphasis"><b>{{.RessetPasswordHash}}</b></p>

    {{if .ResetPasswordPageLink}}
        <p>Click on the following link to go the reset password page - {{.ResetPasswordPageLink}}</p>
    {{end}}

    <p>If you think that this message is a mistake or you need any further help, don't hesitate to contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a>.</p>

    <p>
        Best Regards, <br />
        Gofreta Team
    </p>
{{end}}
`
