package mailsmodels

import (
	"fmt"
	"pec2-backend/utils"
)

func SubscriptionCancellation(email string, creatorName string) {
	subject := "Subject: Confirmation d'annulation d'abonnement OnlyFlick \r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	body := fmt.Sprintf(`
	<div style="background-color: #722ED1; width: 100%%; min-height: 300px; padding: 30px; box-sizing:border-box">
		<table style="background-color: #ffffff; width: 100%%;  min-height: 300px;">
			<tbody>
				<tr>
					<td><h1 style="text-align:center">Confirmation d'annulation d'abonnement</h1></td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 30px;">Votre abonnement à <strong>%s</strong> a bien été annulé.</td>
				</tr>
				<tr>
					<td style="text-align:center; padding-bottom: 20px;">
						<p>Vous ne serez plus débité de 7€ par mois pour cet abonnement.</p>
						<p>Vous pouvez continuer à profiter du contenu jusqu'à la fin de votre période d'abonnement en cours.</p>
					</td>
				</tr>
				<tr>
					<td style="text-align:center; padding-top: 20px; color: #666;">
						<p>Merci d'utiliser OnlyFlick !</p>
					</td>
				</tr>
			</tbody>
		</table>
	</div>
`, creatorName)

	message := []byte(subject + mime + body)

	utils.SendMail(email, message)
}
