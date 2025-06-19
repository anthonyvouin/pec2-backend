package mailsmodels

import (
	"fmt"
	"pec2-backend/utils"
)

func SubscriptionConfirmation(email string, creatorName string) {
	subject := "Subject: Confirmation d'abonnement OnlyFlick \r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
	<div style="background-color: #722ED1; width: 100%%; min-height: 300px; padding: 30px; box-sizing:border-box">
		<table style="background-color: #ffffff; width: 100%%;  min-height: 300px;">
			<tbody>
				<tr>
					<td><h1 style="text-align:center">Confirmation d'abonnement</h1></td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 30px;">Félicitations ! Votre abonnement à <strong>%s</strong> a bien été activé.</td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 20px;">
						<p>Votre abonnement mensuel de 7€ vous donne accès à tout le contenu exclusif de ce créateur.</p>
						<p>Le prochain prélèvement sera effectué dans un mois.</p>
					</td>
				</tr>
				<tr>
					<td style="text-align:center; padding-top: 20px; color: #666;">
						<p>Merci d'utiliser OnlyFlick et de soutenir vos créateurs favoris !</p>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, creatorName)

	message := []byte(subject + mime + body)

	utils.SendMail(email, message)
}
