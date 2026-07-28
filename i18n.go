package main

type Translation struct {
	// App strings
	AppTitle                  string
	AppDescription            string
	DetectDevice              string
	ConnectAndDetect          string
	Scanning                  string
	FlashFirmware             string
	FlashingFirmware          string
	DoNotDisconnect           string
	FlashComplete             string
	KeyboardWillReboot        string
	FlashAnother              string
	NoDeviceFound             string
	FlashFailed               string
	TryAgain                  string
	DetectAgain               string
	ShowLogs                  string
	ViewProgress              string
	ViewLogs                  string
	
	// Device info
	Mode                      string
	ApplicationMode           string
	BootloaderMode            string
	Firmware                  string
	Manufacturer              string
	Product                   string
	
	// Actions
	SelectKeyboard            string
	ChangeKeyboard            string
	BrowseCustomFirmware      string
	Cancel                    string
	Close                     string
	Confirm                   string
	ConfirmSelection          string
	
	// Modals
	SelectYourKeyboard        string
	MultipleModelsDetected    string
	ComingSoon                string
	NotSupported              string
	
	// Warning modal
	Warning                   string
	DangerousOperation        string
	FlashingWrongFirmware     string
	AboutToFlashCustom        string
	AboutToFlashModelPrefix   string
	AboutToFlashModelSuffix   string
	MakeSureMatches           string
	TypePrefix                string
	TypeSuffix                string
	PleaseWaitPrefix          string
	PleaseWaitSuffix          string
	Second                    string
	Seconds                   string
	TextDoesntMatch           string
	ReadyToProceed            string
	FlashNow                  string
	CustomFirmwareConfirm     string
	
	// USB permissions
	USBPermissionsRequired    string
	CopyRules                 string
	ThenPaste                 string
	ContinueAnyway            string
	
	// About
	About                     string
	Copyright                 string
	BuiltWith                 string
	Links                     string
	GitHubRepository          string
	ReportIssue               string
	ViewLicense               string
	
	// Logs
	ConsoleOutput             string
	Copy                      string
	NoLogsAvailable           string
	
	// Errors
	USBPermissionError        string
	InstallUdevRules          string
	UdevRulesRequired         string
	NoDeviceDetected          string
	FlashOperationFailed      string
	CheckLogsForDetails       string
	PermissionDenied          string
	NotCurrentlySupported     string
}

var translations = map[string]Translation{
	"en": {
		AppTitle:                  "DesignedbyGG Updater",
		AppDescription:            "A modern GUI wrapper for SonixFlasher, designed to simplify firmware updates for DesignedbyGG keyboards.",
		DetectDevice:              "Detect Device",
		ConnectAndDetect:          "Connect your keyboard and click detect",
		Scanning:                  "Scanning for devices...",
		FlashFirmware:             "Flash Firmware",
		FlashingFirmware:          "Flashing firmware...",
		DoNotDisconnect:           "Do not disconnect",
		FlashComplete:             "Flash Complete",
		KeyboardWillReboot:        "Your keyboard will reboot automatically",
		FlashAnother:              "Flash Another",
		NoDeviceFound:             "No Device Found",
		FlashFailed:               "Flash Failed",
		TryAgain:                  "Try Again",
		DetectAgain:               "Detect Again",
		ShowLogs:                  "Show Logs",
		ViewProgress:              "View Progress",
		ViewLogs:                  "View Logs",
		Mode:                      "Mode",
		ApplicationMode:           "Application",
		BootloaderMode:            "Bootloader",
		Firmware:                  "Firmware",
		Manufacturer:              "Manufacturer",
		Product:                   "Product",
		SelectKeyboard:            "Select Keyboard",
		ChangeKeyboard:            "Change Keyboard",
		BrowseCustomFirmware:      "Browse for custom firmware...",
		Cancel:                    "Cancel",
		Close:                     "Close",
		Confirm:                   "Confirm",
		ConfirmSelection:          "Confirm Selection",
		SelectYourKeyboard:        "Select Your Keyboard Model",
		MultipleModelsDetected:    "Multiple keyboard models share the same hardware revision. Please select your model:",
		ComingSoon:                "Coming Soon",
		NotSupported:              "This keyboard model is currently not supported!",
		Warning:                   "⚠️ Warning",
		DangerousOperation:        "This is a potentially dangerous operation!",
		FlashingWrongFirmware:     "Flashing wrong firmware can brick your device.",
		AboutToFlashCustom:        "You are about to flash custom firmware. You are responsible for any issues that may occur.",
		AboutToFlashModelPrefix:   "You are about to flash firmware for",
		AboutToFlashModelSuffix:   "Make sure this matches your keyboard model.",
		MakeSureMatches:           "Make sure this matches your keyboard model.",
		TypePrefix:                "Type",
		TypeSuffix:                "to proceed:",
		PleaseWaitPrefix:          "Please wait",
		PleaseWaitSuffix:          "...",
		Second:                    "second",
		Seconds:                   "seconds",
		TextDoesntMatch:           "Text doesn't match",
		ReadyToProceed:            "Ready to proceed",
		FlashNow:                  "Flash Now",
		CustomFirmwareConfirm:     "I like trains",
		USBPermissionsRequired:    "USB Permissions Required",
		CopyRules:                 "Copy rules",
		ThenPaste:                 "then:",
		ContinueAnyway:            "Continue anyway",
		About:                     "About",
		Copyright:                 "Copyright",
		BuiltWith:                 "Built With",
		Links:                     "Links",
		GitHubRepository:          "GitHub Repository",
		ReportIssue:               "Report an Issue",
		ViewLicense:               "View License",
		ConsoleOutput:             "Console Output",
		Copy:                      "Copy",
		NoLogsAvailable:           "No logs available",
		USBPermissionError:        "USB Permission Error: Please install udev rules",
		InstallUdevRules:          "USB Permission Error - Udev rules required",
		UdevRulesRequired:         "USB Permission Error - Udev rules required",
		NoDeviceDetected:          "No device detected",
		FlashOperationFailed:      "Flash operation failed",
		CheckLogsForDetails:       "Flash failed - check logs for details",
		PermissionDenied:          "Permission denied",
		NotCurrentlySupported:     "This keyboard model is currently not supported!",
	},
	"fr": {
		AppTitle:                  "DesignedbyGG Updater",
		AppDescription:            "Une interface graphique moderne pour SonixFlasher, conçue pour simplifier les mises à jour de firmware pour les claviers DesignedbyGG.",
		DetectDevice:              "Détecter l'appareil",
		ConnectAndDetect:          "Connectez votre clavier et cliquez sur détecter",
		Scanning:                  "Recherche d'appareils...",
		FlashFirmware:             "Flasher le Firmware",
		FlashingFirmware:          "Flashage du firmware...",
		DoNotDisconnect:           "Ne pas déconnecter",
		FlashComplete:             "Flashage Terminé",
		KeyboardWillReboot:        "Votre clavier redémarrera automatiquement",
		FlashAnother:              "Flasher un autre",
		NoDeviceFound:             "Aucun Appareil Trouvé",
		FlashFailed:               "Échec du Flashage",
		TryAgain:                  "Réessayer",
		DetectAgain:               "Détecter à nouveau",
		ShowLogs:                  "Afficher les Logs",
		ViewProgress:              "Voir la Progression",
		ViewLogs:                  "Voir les Logs",
		Mode:                      "Mode",
		ApplicationMode:           "Application",
		BootloaderMode:            "Bootloader",
		Firmware:                  "Firmware",
		Manufacturer:              "Fabricant",
		Product:                   "Produit",
		SelectKeyboard:            "Sélectionner le Clavier",
		ChangeKeyboard:            "Changer de Clavier",
		BrowseCustomFirmware:      "Parcourir pour un firmware personnalisé...",
		Cancel:                    "Annuler",
		Close:                     "Fermer",
		Confirm:                   "Confirmer",
		ConfirmSelection:          "Confirmer la Sélection",
		SelectYourKeyboard:        "Sélectionnez Votre Modèle de Clavier",
		MultipleModelsDetected:    "Plusieurs modèles de clavier partagent la même révision matérielle. Veuillez sélectionner votre modèle:",
		ComingSoon:                "Bientôt Disponible",
		NotSupported:              "Ce modèle de clavier n'est actuellement pas pris en charge!",
		Warning:                   "⚠️ Attention",
		DangerousOperation:        "Ceci est une opération potentiellement dangereuse!",
		FlashingWrongFirmware:     "Flasher un mauvais firmware peut endommager votre appareil.",
		AboutToFlashCustom:        "Vous êtes sur le point de flasher un firmware personnalisé. Vous êtes responsable de tout problème qui pourrait survenir.",
		AboutToFlashModelPrefix:   "Vous êtes sur le point de flasher le firmware pour",
		AboutToFlashModelSuffix:   "Assurez-vous que cela correspond à votre modèle de clavier.",
		MakeSureMatches:           "Assurez-vous que cela correspond à votre modèle de clavier.",
		TypePrefix:                "Tapez",
		TypeSuffix:                "pour continuer:",
		PleaseWaitPrefix:          "Veuillez patienter",
		PleaseWaitSuffix:          "...",
		Second:                    "seconde",
		Seconds:                   "secondes",
		TextDoesntMatch:           "Le texte ne correspond pas",
		ReadyToProceed:            "Prêt à continuer",
		FlashNow:                  "Flasher Maintenant",
		CustomFirmwareConfirm:     "Je suis une baguette",
		USBPermissionsRequired:    "Permissions USB Requises",
		CopyRules:                 "Copier les règles",
		ThenPaste:                 "puis:",
		ContinueAnyway:            "Continuer quand même",
		About:                     "À Propos",
		Copyright:                 "Droits d'auteur",
		BuiltWith:                 "Construit avec",
		Links:                     "Liens",
		GitHubRepository:          "Dépôt GitHub",
		ReportIssue:               "Signaler un Problème",
		ViewLicense:               "Voir la Licence",
		ConsoleOutput:             "Sortie Console",
		Copy:                      "Copier",
		NoLogsAvailable:           "Aucun log disponible",
		USBPermissionError:        "Erreur de permission USB: Veuillez installer les règles udev",
		InstallUdevRules:          "Erreur de permission USB - Règles udev requises",
		UdevRulesRequired:         "Erreur de permission USB - Règles udev requises",
		NoDeviceDetected:          "Aucun appareil détecté",
		FlashOperationFailed:      "Échec de l'opération de flashage",
		CheckLogsForDetails:       "Échec du flashage - vérifiez les logs pour plus de détails",
		PermissionDenied:          "Permission refusée",
		NotCurrentlySupported:     "Ce modèle de clavier n'est actuellement pas pris en charge!",
	},
}

func (a *App) GetTranslations(lang string) Translation {
	if t, ok := translations[lang]; ok {
		return t
	}
	return translations["en"]
}

func (a *App) GetAvailableLanguages() []map[string]string {
	return []map[string]string{
		{"code": "en", "name": "English", "flag": "🍔"},
		{"code": "fr", "name": "Français", "flag": "🥐"},
	}
}