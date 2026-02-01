package service

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v3"

	"silsilah-keluarga/internal/config"
)

type EmailService interface {
	SendRegistrationEmail(ctx context.Context, toEmail, fullName string) error
	SendEmailVerification(ctx context.Context, toEmail, fullName, verificationToken string) error
	SendPasswordResetEmail(ctx context.Context, toEmail, fullName, resetToken string) error
}

type emailService struct {
	client *resend.Client
	config *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	client := resend.NewClient(cfg.ResendAPIKey)
	return &emailService{
		client: client,
		config: cfg,
	}
}

func (s *emailService) SendRegistrationEmail(ctx context.Context, toEmail, fullName string) error {
	subject := "Selamat Datang di Silsilah Keluarga!"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Selamat Datang di Silsilah Keluarga</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f9fafb;">

	<!-- Header -->
	<div style="background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
		<h1 style="color: #ffffff; margin: 0; font-size: 28px;">
			Silsilah Keluarga
		</h1>
		<p style="color: #d1fae5; margin: 10px 0 0 0; font-size: 16px;">
			Menghubungkan Generasi, Menjaga Kisah Keluarga Kita
		</p>
	</div>

	<!-- Content -->
	<div style="background-color: #ffffff; padding: 30px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">

		<h2 style="color: #111827; margin-top: 0;">
			Halo, %s!
		</h2>

		<p>
			Terima kasih telah bergabung dengan <strong>Silsilah Keluarga</strong>.
			Bersama, kita membangun silsilah keluarga kita untuk menghubungkan generasi dan melestarikan kisah keluarga kita.
		</p>

		<div style="background-color: #f3f4f6; padding: 20px; border-radius: 8px; margin: 20px 0;">
			<h3 style="margin-top: 0; color: #111827;">
				Informasi Akun
			</h3>
			<div style="margin-bottom: 10px;">
				<strong>Email:</strong> %s
			</div>
			<div>
				<strong>Nama:</strong> %s
			</div>
		</div>

		<p>
			Akun Anda telah berhasil dibuat. Sekarang Anda dapat mulai membangun pohon keluarga, 
			menghubungkan kerabat, dan melestarikan sejarah keluarga Anda.
		</p>

		<!-- Button -->
		<div style="text-align: center; margin: 30px 0;">
			<a href="%s"
			   style="background-color: #10b981; color: #ffffff; padding: 14px 28px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">
				Masuk ke Akun Anda
			</a>
		</div>

		<p style="font-size: 14px; color: #6b7280;">
			Jika Anda memiliki pertanyaan atau membutuhkan bantuan, jangan ragu untuk menghubungi tim dukungan kami.
		</p>

		<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 30px 0;">

		<p style="font-size: 14px; color: #6b7280;">
			Salam hangat,<br>
			<strong>Tim Silsilah Keluarga</strong>
		</p>
	</div>

	<!-- Footer -->
	<div style="text-align: center; margin-top: 20px; font-size: 12px; color: #9ca3af;">
		<p>
			Dibuat oleh
			<a href="https://zamili.dev" style="color: #9ca3af; text-decoration: underline;">
				Nestor Zamili
			</a>
		</p>
		<p>
			© 2026 Silsilah Keluarga. Hak cipta dilindungi.
		</p>
	</div>

</body>
</html>`, fullName, toEmail, fullName, fmt.Sprintf("http://%s/login", s.config.Domain))

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Silsilah Keluarga <%s>", s.config.FromEmail),
		To:      []string{toEmail},
		Html:    html,
		Subject: subject,
	}

	_, err := s.client.Emails.Send(params)
	return err
}

func (s *emailService) SendEmailVerification(ctx context.Context, toEmail, fullName, verificationToken string) error {
	subject := "Verifikasi Email - Silsilah Keluarga"
	verificationLink := fmt.Sprintf("https://%s/verify-email?token=%s", s.config.Domain, verificationToken)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Verifikasi Email - Silsilah Keluarga</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f9fafb;">

	<!-- Header -->
	<div style="background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
		<h1 style="color: #ffffff; margin: 0; font-size: 28px;">
			Silsilah Keluarga
		</h1>
		<p style="color: #d1fae5; margin: 10px 0 0 0; font-size: 16px;">
			Menghubungkan Generasi, Menjaga Kisah Keluarga Kita
		</p>
	</div>

	<!-- Content -->
	<div style="background-color: #ffffff; padding: 30px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">

		<h2 style="color: #111827; margin-top: 0;">
			Halo, %s!
		</h2>

		<p>
			Terima kasih telah bergabung dengan <strong>Silsilah Keluarga</strong>.
			Bersama, kita membangun silsilah keluarga kita untuk menghubungkan generasi dan melestarikan kisah keluarga kita.
		</p>

		<div style="background-color: #eff6ff; padding: 20px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #3b82f6;">
			<h3 style="margin-top: 0; color: #1e40af;">
				Langkah Verifikasi
			</h3>
			<p style="margin-bottom: 0;">
				Klik tombol di bawah ini untuk memverifikasi alamat email Anda.  
				Tautan ini akan kedaluwarsa dalam <strong>24 jam</strong>.
			</p>
		</div>

		<!-- Button -->
		<div style="text-align: center; margin: 30px 0;">
			<a href="%s"
			   style="background-color: #3b82f6; color: #ffffff; padding: 14px 28px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">
				Verifikasi Email Saya
			</a>
		</div>

		<!-- Fallback Link -->
		<p style="font-size: 14px; color: #6b7280;">
			Jika tombol di atas tidak berfungsi, salin dan tempel tautan berikut ke browser Anda:
			<br>
			<a href="%s" style="color: #3b82f6; word-break: break-all;">
				%s
			</a>
		</p>

		<!-- Security Note -->
		<p style="font-size: 14px; color: #6b7280;">
			Jika Anda tidak merasa mendaftar di Silsilah Keluarga, abaikan email ini.  
			Demi keamanan, jangan pernah membagikan tautan verifikasi kepada siapa pun.
		</p>

		<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 30px 0;">

		<p style="font-size: 14px; color: #6b7280;">
			Salam hangat,<br>
			<strong>Tim Silsilah Keluarga</strong>
		</p>
	</div>

	<!-- Footer -->
	<div style="text-align: center; margin-top: 20px; font-size: 12px; color: #9ca3af;">
		<p>
			Dibuat oleh
			<a href="https://zamili.dev" style="color: #9ca3af; text-decoration: underline;">
				Nestor Zamili
			</a>
		</p>
		<p>
			© 2026 Silsilah Keluarga. Hak cipta dilindungi.
		</p>
	</div>

</body>
</html>`, fullName, verificationLink, verificationLink, verificationLink)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Silsilah Keluarga <%s>", s.config.FromEmail),
		To:      []string{toEmail},
		Html:    html,
		Subject: subject,
	}

	_, err := s.client.Emails.Send(params)
	return err
}

func (s *emailService) SendPasswordResetEmail(ctx context.Context, toEmail, fullName, resetToken string) error {
	subject := "Permintaan Reset Kata Sandi - Silsilah Keluarga"
	resetLink := fmt.Sprintf("https://%s/reset-password?token=%s", s.config.Domain, resetToken)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Reset Kata Sandi - Silsilah Keluarga</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f9fafb;">

	<!-- Header -->
	<div style="background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
		<h1 style="color: #ffffff; margin: 0; font-size: 28px;">
			Silsilah Keluarga
		</h1>
		<p style="color: #d1fae5; margin: 10px 0 0 0; font-size: 16px;">
			Menghubungkan Generasi, Menjaga Kisah Keluarga Kita
		</p>
	</div>

	<!-- Content -->
	<div style="background-color: #ffffff; padding: 30px; border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px;">

		<h2 style="color: #111827; margin-top: 0;">
			Halo, %s!
		</h2>

		<p>
			Kami menerima permintaan untuk mereset kata sandi akun Silsilah Keluarga Anda.
		</p>

		<div style="background-color: #fffbeb; padding: 20px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #f59e0b;">
			<h3 style="margin-top: 0; color: #92400e;">
				Langkah Reset Kata Sandi
			</h3>
			<p style="margin-bottom: 0;">
				Klik tombol di bawah ini untuk mereset kata sandi Anda. 
				Tautan ini akan kedaluwarsa dalam <strong>1 jam</strong>.
			</p>
		</div>

		<!-- Button -->
		<div style="text-align: center; margin: 30px 0;">
			<a href="%s"
			   style="background-color: #f59e0b; color: #ffffff; padding: 14px 28px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">
				Reset Kata Sandi Saya
			</a>
		</div>

		<!-- Fallback Link -->
		<p style="font-size: 14px; color: #6b7280;">
			Jika tombol di atas tidak berfungsi, salin dan tempel tautan berikut ke browser Anda:
			<br>
			<a href="%s" style="color: #f59e0b; word-break: break-all;">
				%s
			</a>
		</p>

		<!-- Security Note -->
		<p style="font-size: 14px; color: #6b7280;">
			Jika Anda tidak meminta reset kata sandi ini, harap abaikan email ini 
			atau hubungi tim dukungan kami jika Anda memiliki kekhawatiran.
			Demi keamanan, jangan pernah membagikan tautan reset kepada siapa pun.
		</p>

		<hr style="border: none; border-top: 1px solid #e5e7eb; margin: 30px 0;">

		<p style="font-size: 14px; color: #6b7280;">
			Salam hangat,<br>
			<strong>Tim Silsilah Keluarga</strong>
		</p>
	</div>

	<!-- Footer -->
	<div style="text-align: center; margin-top: 20px; font-size: 12px; color: #9ca3af;">
		<p>
			Dibuat oleh
			<a href="https://zamili.dev" style="color: #9ca3af; text-decoration: underline;">
				Nestor Zamili
			</a>
		</p>
		<p>
			© 2026 Silsilah Keluarga. Hak cipta dilindungi.
		</p>
	</div>

</body>
</html>`, fullName, resetLink, resetLink, resetLink)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("Silsilah Keluarga <%s>", s.config.FromEmail),
		To:      []string{toEmail},
		Html:    html,
		Subject: subject,
	}

	_, err := s.client.Emails.Send(params)
	return err
}
