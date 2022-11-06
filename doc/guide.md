# Boursière

## Algorithme

Le prix des différentes bières est mis à jour toutes les 15 minutes. Notons que ce chiffre est *hard-codé* dans le code source, autant du côté du serveur que du client. Si, par hasard, vous souhaitiez le modifier, référez-vous à la [section « Personnalisation »](#personnalisation). Par la suite, nous parlerons de « période » pour désigner cette durée.

Contrairement à ce que l'on pourrait penser, le prix d'une bière est totalement indépendant des autres bières. Il ne dépend que des ventes de la bière en question.

Chaque bière possède ses paramètre spécifiques :

* purchasePrice : le prix d'achat pour une unité.
* incrCoef (*increase coeficient*) : le coefficient d'augmentation du prix de la bière (voir explications ci-dessous).
* decrCoef (*decrease coeficient*) : le coefficient de diminution du prix de la bière.
* minCoef (*minimum coeficient*) : le coefficient du prix minimum de la bière. Ce facteur multiplié par le prix d'achat représente le prix minimum de la bière en-dessous duquel le prix de vente ne descendra jamais.
* maxCoef (*maximum coeficient*) : le coefficient du prix maximum de la bière. À l'instar du paramètre précédent, ce facteur détermine le prix maximum de la bière.

Au démarrage, le prix de vente de chaque bière (sellingPrice) est fixé au prix d'achat (purchasePrice). Ensuite, à chaque période, le nouveau prix d'une bière est calculé comme suit :

1. Soit une bière **b**.
2. Soit **δ** = **b**.soldQuantity − **b**.previousSoldQuantity. Il s'agit du nombre de bières vendues en plus (ou en moins) durant la période actuelle par rapport à la période précédente.
3. Si **δ** est nul, le prix de **b** ne varie pas.
4. Sinon, si **δ** est positif, **δ** × **b**.incrCoef est additionné au prix actuel.
5. Sinon, si **δ** est négatif, **δ** × **b**.decrCoef est additionné au prix actuel. Notons bien que dans ce cas, le prix diminue, car **δ** est négatif.
6. Enfin, le nouveau prix est borné, tel que **b**.minCoef × **b**.purchasePrice < **b**.sellingPrice < **b**.maxCoef × **b**.purchasePrice.

Quelques remarques par rapport à cet algorithme :

* En définissant des valeurs différentes pour les coefficients d'augmentation et de diminution des prix, il est possible de définir la tendance du prix d'une bière. Par exemple, si incrCoef > decrCoef, le prix aura plutôt tendance à augmenter qu'à diminuer.
* Si nous utilisons 1 pour le coefficient minimum (minCoef) d'une bière, cette dernière ne sera jamais vendue à perte. Cependant, vendre une bière à perte n'est pas forcément une mauvaise chose. N'oubliez pas que la majeure partie des bénéfices vient des entrées et que les gens garderont un meilleur souvenir de la soirée s'ils parviennent à faire de bonnes affaires.
* Enfin, le [template des bières](./beers.ods) contient une feuille qui permet de simuler l'algorithme avec des données fictives. Par ailleurs, ce template contient les paramètres utilisés en 2019 et peut vous servir de référence.

## Démarrage

Une fois le serveur démarré — `go run .` en développement, cf. [Dockerfile](../Dockerfile) pour la production — connectez-vous sur la [page d'administration](#page-dadministration) avec l'utilisateur `admin` et le mot de passe `boursière`. **Changez immédiatement le mot de passe**, car il est publiquement disponible sur ce dépôt. Ensuite, créez les différents utilisateurs pour l'événement et importez les bières dans le système.

## Utilisation

### Page principale

> `/index.html`

Il s'agit de la page d'accueil, disponible à tous, qui présente les différentes bières disponibles. Il est possible de trier les bières par bar en suffixant `#1`, `#2`, `#3`, etc. à l'URL.

L'ordre des bières dans la liste est le même que celui du fichier qui est importé dans le système (cf. [« Page d'administration »](#page-dadministration)).

Les prix et quantités se mettent à jour en temps réel. Cela est rendu possible par l'utilisation de [*server-sent events (SSE)*](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events).

Un son est joué à chaque nouvelle période. Cependant, par défaut, les navigateurs bloquent les fichiers audio qui sont joués automatiquement par une page web. Si vous souhaitez donc entendre le son, il faut donner cette autorisation dans votre navigateur.

### Page de commande

> `/order.html`

Cette page permet d'inscrire les commandes lors de la soirée. Il faut se connecter pour y accéder mais pas nécessairement en tant qu'administrateur. Comme pour la page principale, il est possible de trier les bières par bar, ce de la même manière.

Le prix de chaque commande est calculé automatiquement, et ce sur base des données mises à jour en temps réel. Il y a également un historique local qui permet de voir les 5 dernières commandes ainsi que d'annuler une commande erronée.

Enfin, il est possible à tout moment de passer des « commandes négatives » afin d'augmenter le stock de bières dans le système. Notons d'ailleurs que c'est de cette manière qu'est implémenté l'historique.

### Page d'administration

> `/admin.html`

C'est sur cette page que vous pouvez importer les bières dans le système et créer, modifier ou supprimer des utilisateurs. Elle n'est accessible qu'aux administrateurs.

Attention, **importer des bières réinitialise l'intégralité de l'événement** : toutes les commandes effectuées seront supprimées et les prix seront remis à zéro. Cet effet est volontaire, car cela permet de réinitialiser la boursière quand c'est nécessaire, mais soyez-en conscients.

### Page de gestion des entrées

> `/entries.html`

Cette page permet de gérer la vente et revente des préventes durant la soirée.
Les entrées sont remise à zéro en même temps que le prix des bières et les stocks via l'importation d'un
nouveau fichier .csv via le panel admin.

Le prix d'une prévente est actuellement hardcodé sur le client (il n'est pas présent sur le serveur). Si vous devez
le modifier, il faudra le modifer sur `EntriesCard.tsx` et `AdminStats.txt`.

## Personnalisation

Il est possible de personnaliser l'apparence du site web en modifiant les fichiers CSS. Vous pouvez notamment changer la couleur principale ou la taille de la police.

Si vous souhaitez modifier le son de changement de période, il suffit de modifier le fichier audio dans le code source du client web. Comme pour toute application web, il est fortement conseillé de choisir un format audio compressé (Ogg Vorbis idéalement) pour limiter l'utilisation de la bande passante.

Ensuite, pour changer la durée d'une période, il faut à la fois modifier le fichier "main.go" côté serveur et "BeerTimer.tsx" côté client.

Finalement, portez une attention au prix de la prévente, comme expliqué dans la section `Page de gestion des entrées`
