package emails

const Layout = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <title>Reset Password</title>

    <style>
        body, html {
            padding: 0;
            margin: 0;
            border: 0;
        }
        body, html,
        .global-wrapper {
            color: #4d5961;
            background: #f3f6f8;
            font-size: 14px;
            line-height: 24px;
            font-weight: normal;
            font-family: Arial, Helvetica, sans-serif;
        }
        .global-wrapper {
            display: block;
            width: 100%;
            height: 100%;
            padding: 30px 0;
        }
        .wrapper {
            width: 500px;
            max-width: 90%;
            margin: 0 auto;
            font-size: inherit;
            font-family: inherit;
            line-height: inherit;
        }
        p {
            display: block;
            margin: 10px 0;
        }
        table {
            color: inherit;
            font-size: inherit;
            font-weight: inherit;
            font-family: inherit;
            line-height: inherit;
            width: 100%;
            max-width: 100%;
            border-spacing: 0;
            border-collapse: collapse;
        }
        td, th {
            border: 0;
            color: inherit;
            vertical-align: middle;
        }
        a {
            color: inherit !important;
            text-decoration: underline !important;
        }
        a:hover {
            color: inherit !important;
            text-decoration: none !important;
        }
        .btn {
            display: inline-block;
            vertical-align: top;
            border: 0;
            color: #fff !important;
            background: #00d9a4 !important;
            text-decoration: none !important;
            line-height: 42px;
            min-width: 130px;
            text-align: center;
            padding: 0 30px;
            margin: 10px 0;
            font-weight: bold;
            border-radius: 3px;
        }
        .btn:hover {
            color: #fff !important;
            background: #03ca99 !important;
        }
        .btn:active {
            color: #fff !important;
            background: #01c192 !important;
        }
        .hint {
            font-size: 12px;
            line-height: 14px;
        }
        .shadowed {
            background: #fff;
            border-radius: 3px;
            box-shadow: 0px 2px 10px 0px rgba(44, 75, 137, 0.12);
        }
        .emphasis {
            padding: 15px;
            background: #f3f6f8;
            border-radius: 3px;
        }
        .text-center {
            text-align: center;
        }
        .text-left {
            text-align: left;
        }
        .text-right {
            text-align: right;
        }

        /* Header */
        .header {
            display: block;
            padding: 20px;
            border: 0;
            text-align: center;
        }
        .logo img {
            vertical-align: middle;
            display: inline-block;
        }

        /* Content */
        .content {
            display: block;
            padding: 15px 30px;
            background: #fff;
            border-top: 0px;
            border-radius: 3px;
        }

        /* Footer */
        .footer {
            display: block;
            padding: 20px 0;
            margin: 0;
            text-align: left;
            font-size: 12px;
            line-height: 14px;
            color: #a7b2c9;
        }
        .footer p {
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="global-wrapper">
        <div class="wrapper">
            <div class="header">
                <h1>Gofreta</h1>
            </div>
        </div>

        <div class="wrapper shadowed">
            <div class="content">
                {{template "content" .}}
            </div>
        </div>

        <div class="wrapper">
            <div class="footer">
                <table style="width: 100%;">
                    <tr>
                        <td style="width: 50%; text-align: left;">
                            <p>Gofreta CMS</p>
                        </td>
                        <td style="width: 50%; text-align: right;">
                            <a href="mailto:{{.SupportEmail}}" class="social-link">{{.SupportEmail}}</a>
                        </td>
                    </tr>
                </table>
            </div>
        </div>
    </div>
</body>
</html>
`
