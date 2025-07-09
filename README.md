# OnlyFlick - Backend API Go

## Explication du Projet
OnlyFlick est une plateforme sociale innovante dédiée aux amoureux des animaux, permettant aux utilisateurs de créer, partager et interagir avec du contenu animalier. Le projet vise à offrir une expérience utilisateur complète avec des fonctionnalités de partage de photos et vidéos d'animaux, d'échange de conseils sur le soin animalier, de messagerie privée entre passionnés, et un système d'abonnement pour suivre les créateurs de contenu spécialisés. La plateforme propose également des fonctionnalités de modération pour assurer un contenu respectueux et des options de monétisation pour les créateurs partageant du contenu éducatif ou divertissant sur le monde animal.

## Lien vers le Projet
Vous pouvez accéder à la partie utilisateur en ligne à l'adresse suivante : [OnlyFlick (Frontend)](https://onlyflick.akiagaming.fr)

Vous pouvez accéder à la partie API en ligne à l'adresse suivante : [OnlyFlick (Backend)](https://api.onlyflick.akiagaming.fr)

## Télécharger l'Application (Android)

Pour télécharger le projet sur votre téléphone, vous pouvez accéder au dépôt GitHub : [OnlyFlick GitHub](https://github.com/anthonyvouin/pec2-frontend)

Pour installer l'application mobile :
1. Rendez-vous dans la section "Releases" du dépôt GitHub
2. Téléchargez la dernière version stable (.apk pour Android)
3. Suivez les instructions d'installation pour votre appareil

Notez que vous devrez peut-être autoriser l'installation d'applications provenant de sources inconnues dans les paramètres de votre téléphone.

## Contributions
- **Anthony Vouin** - github : anthonyvouin : Développeur Full Stack
- **Charline Royer** - github : akia-web : Développeur Full Stack  
- **Matthias Faucon** - github : matthiasfaucon : Développeur Full Stack

## Technologies Utilisées (Backend)
- **Go** : Langage de programmation principal (version 1.24)
- **Gin** : Framework web HTTP
- **PostgreSQL** : Base de données relationnelle
- **GORM** : ORM pour Go
- **Stripe** : Plateforme de paiement pour les abonnements
- **Cloudinary** : Gestion et stockage des médias
- **Swagger** : Documentation API automatique
- **Docker** : Conteneurisation de l'application
- **NeonDB** : Base de données PostgreSQL cloud (optionnel)

## Fonctionnalités

### Infrastructure et DevOps
- **Mise en place de Kubernetes** - Charline Royer
- **Mise en place de Grafana** - Matthias Faucon
- **Dockeriser le Flutter** - Charline Royer
- **Dockeriser le web de Flutter** - Charline Royer
- **Dockeriser le Golang** - Charline Royer
- **CI Flutter** - Matthias Faucon
- **CI Golang** - Matthias Faucon
- **CD web Flutter** - Matthias Faucon
- **CD APK Flutter** - Matthias Faucon
- **CD Golang** - Matthias Faucon

### Gestion des Utilisateurs
- **Éditer son profil** - Charline Royer
- **Changer mot de passe** - Anthony Vouin
- **Page de profil** - Anthony Vouin
- **Fonctionnalité mot de passe oublié Flutter** - Charline Royer
- **Formulaire inscription / connexion** - Charline Royer
- **Page création de compte (register)** - Anthony Vouin
- **Validation du compte** - Charline Royer

### Fonctionnalités Sociales
- **Création de posts** - Matthias Faucon
- **Mise en place des likes sur les posts** - Matthias Faucon
- **Mise en place des commentaires en SSE sur les posts** - Matthias Faucon
- **Mise en place des signalements sur les posts** - Matthias Faucon
- **Fonctionnalité des messages privées** - Anthony Vouin
- **Fonctionnalité de suivi des autres utilisateurs** - Anthony Vouin / Charline Royer

### Créateurs de Contenu
- **Demande pour devenir créateur de contenus** - Anthony Vouin / Charline Royer
- **Vérification du SIRET + SIREN** - Anthony Vouin
- **Paiements par Stripe (abonnement)** - Anthony Vouin (backend) / Charline Royer (frontend)

### Administration
- **Statistiques (partie administrateur)** - Anthony Vouin
- **Statistiques (partie utilisateur et créateur de contenus)** - Charline Royer
- **Administration/gestion des données (partie administrateur)** - Anthony Vouin
- **Activer/Désactiver fonctionnalité abonnement pour un utilisateur spécifique** - Matthias Faucon
- **Activer/Désactiver fonctionnalité message privé pour un utilisateur spécifique** - Matthias Faucon
- **Activer/Désactiver fonctionnalité commentaires pour un utilisateur spécifique** - Matthias Faucon

### Fonctionnalités Techniques
- **Gestion du thème** - Matthias Faucon
- **Mise en place des logs structurés (Golang)** - Anthony Vouin

## Rôles et Permissions

Dans ce projet, il existe trois rôles principaux : **ADMIN**, **CONTENT_CREATOR** (créateur de contenu), et **USER** (utilisateur classique). Voici un aperçu des permissions associées à chaque rôle :

### ADMIN :
- Accès complet aux fonctionnalités administratives
- Gestion des utilisateurs et des créateurs de contenu
- Accès aux statistiques globales

### CONTENT_CREATOR (Créateur de contenu) :
- Gestion des abonnements et revenus
- Créer des posts privés

### USER (Utilisateur classique) :
- Création et gestion de posts
- Accès aux statistiques de leurs contenus
- Consultation et interaction avec les contenus
- Gestion de son profil personnel
- Système de likes et commentaires
- Messagerie privée
- Abonnements aux créateurs de contenu

#### Note : Les fonctionnalités proposées aux utilisateurs classiques, sont également disponibles pour les créateurs de contenu, mais avec des permissions supplémentaires pour la gestion de leurs propres contenus.

## Comment Lancer le Projet Backend

## Prérequis

- **Go** : Version 1.21 ou supérieure
- **PostgreSQL** : Instance de base de données (ou NeonDB)
- **Stripe CLI** : Pour tester les webhooks (optionnel)

## Installation

1. **Cloner le repository :**

```bash
git clone https://github.com/anthonyvouin/pec2-backend
cd pec2-backend
```

2. **Installer les dépendances :**

```bash
go mod tidy
```

3. **Lancer l'application :**

```bash
go run main.go
```

L'API sera disponible sur `http://localhost:8080`

## Documentation API (Swagger)

### Générer la documentation Swagger :

```bash
swag init
```

### Accéder à la documentation :
```bash
http://localhost:8080/swagger/index.html
```

## Tests

### Lancer les tests pour une fonctionnalité spécifique :
```bash
go test ./handlers/auth -v
```

### Lancer tous les tests :
```bash
go test ./...
```

## Stripe (Développement)

### Écouter les webhooks Stripe en local :
```bash
stripe listen --forward-to http://localhost:8080/stripe/webhook
```

## Docker (Backend)

### Build et lancement avec Docker :
```bash
# Build de l'image
docker build -t onlyflick-backend .

# Lancement avec docker-compose
docker-compose up -d
```

### Variables d'environnement Docker :
Assurez-vous de configurer les variables d'environnement dans le `docker-compose.yaml` ou via un fichier `.env`.

## Base de Données

### Migrations automatiques :
L'application utilise GORM AutoMigrate pour créer automatiquement les tables au démarrage.

### Connexion à la base de données :
- **Local** : PostgreSQL traditionnel
- **Production** : NeonDB (PostgreSQL cloud)