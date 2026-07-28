export namespace main {
	
	export class AppModeDevice {
	    name: string;
	    description: string;
	    firmwarePath: string;
	    bootloaderVID: number;
	    bootloaderPID: number;
	    bcdDevice: number;
	
	    static createFrom(source: any = {}) {
	        return new AppModeDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.firmwarePath = source["firmwarePath"];
	        this.bootloaderVID = source["bootloaderVID"];
	        this.bootloaderPID = source["bootloaderPID"];
	        this.bcdDevice = source["bcdDevice"];
	    }
	}
	export class Device {
	    vid: string;
	    pid: string;
	    name: string;
	    manufacturer: string;
	    product: string;
	    serialNumber: string;
	    path: string;
	    isBootloader: boolean;
	    firmwarePath: string;
	    candidates?: string;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vid = source["vid"];
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.manufacturer = source["manufacturer"];
	        this.product = source["product"];
	        this.serialNumber = source["serialNumber"];
	        this.path = source["path"];
	        this.isBootloader = source["isBootloader"];
	        this.firmwarePath = source["firmwarePath"];
	        this.candidates = source["candidates"];
	    }
	}
	export class FlashResult {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new FlashResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class Translation {
	    AppTitle: string;
	    AppDescription: string;
	    DetectDevice: string;
	    ConnectAndDetect: string;
	    Scanning: string;
	    FlashFirmware: string;
	    FlashingFirmware: string;
	    DoNotDisconnect: string;
	    FlashComplete: string;
	    KeyboardWillReboot: string;
	    FlashAnother: string;
	    NoDeviceFound: string;
	    FlashFailed: string;
	    TryAgain: string;
	    DetectAgain: string;
	    ShowLogs: string;
	    ViewProgress: string;
	    ViewLogs: string;
	    Mode: string;
	    ApplicationMode: string;
	    BootloaderMode: string;
	    Firmware: string;
	    Manufacturer: string;
	    Product: string;
	    SelectKeyboard: string;
	    ChangeKeyboard: string;
	    BrowseCustomFirmware: string;
	    Cancel: string;
	    Close: string;
	    Confirm: string;
	    ConfirmSelection: string;
	    SelectYourKeyboard: string;
	    MultipleModelsDetected: string;
	    ComingSoon: string;
	    NotSupported: string;
	    Warning: string;
	    DangerousOperation: string;
	    FlashingWrongFirmware: string;
	    AboutToFlashCustom: string;
	    AboutToFlashModelPrefix: string;
	    AboutToFlashModelSuffix: string;
	    MakeSureMatches: string;
	    TypePrefix: string;
	    TypeSuffix: string;
	    PleaseWaitPrefix: string;
	    PleaseWaitSuffix: string;
	    Second: string;
	    Seconds: string;
	    TextDoesntMatch: string;
	    ReadyToProceed: string;
	    FlashNow: string;
	    CustomFirmwareConfirm: string;
	    USBPermissionsRequired: string;
	    CopyRules: string;
	    ThenPaste: string;
	    ContinueAnyway: string;
	    About: string;
	    Copyright: string;
	    BuiltWith: string;
	    Links: string;
	    GitHubRepository: string;
	    ReportIssue: string;
	    ViewLicense: string;
	    ConsoleOutput: string;
	    Copy: string;
	    NoLogsAvailable: string;
	    USBPermissionError: string;
	    InstallUdevRules: string;
	    UdevRulesRequired: string;
	    NoDeviceDetected: string;
	    FlashOperationFailed: string;
	    CheckLogsForDetails: string;
	    PermissionDenied: string;
	    NotCurrentlySupported: string;
	
	    static createFrom(source: any = {}) {
	        return new Translation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AppTitle = source["AppTitle"];
	        this.AppDescription = source["AppDescription"];
	        this.DetectDevice = source["DetectDevice"];
	        this.ConnectAndDetect = source["ConnectAndDetect"];
	        this.Scanning = source["Scanning"];
	        this.FlashFirmware = source["FlashFirmware"];
	        this.FlashingFirmware = source["FlashingFirmware"];
	        this.DoNotDisconnect = source["DoNotDisconnect"];
	        this.FlashComplete = source["FlashComplete"];
	        this.KeyboardWillReboot = source["KeyboardWillReboot"];
	        this.FlashAnother = source["FlashAnother"];
	        this.NoDeviceFound = source["NoDeviceFound"];
	        this.FlashFailed = source["FlashFailed"];
	        this.TryAgain = source["TryAgain"];
	        this.DetectAgain = source["DetectAgain"];
	        this.ShowLogs = source["ShowLogs"];
	        this.ViewProgress = source["ViewProgress"];
	        this.ViewLogs = source["ViewLogs"];
	        this.Mode = source["Mode"];
	        this.ApplicationMode = source["ApplicationMode"];
	        this.BootloaderMode = source["BootloaderMode"];
	        this.Firmware = source["Firmware"];
	        this.Manufacturer = source["Manufacturer"];
	        this.Product = source["Product"];
	        this.SelectKeyboard = source["SelectKeyboard"];
	        this.ChangeKeyboard = source["ChangeKeyboard"];
	        this.BrowseCustomFirmware = source["BrowseCustomFirmware"];
	        this.Cancel = source["Cancel"];
	        this.Close = source["Close"];
	        this.Confirm = source["Confirm"];
	        this.ConfirmSelection = source["ConfirmSelection"];
	        this.SelectYourKeyboard = source["SelectYourKeyboard"];
	        this.MultipleModelsDetected = source["MultipleModelsDetected"];
	        this.ComingSoon = source["ComingSoon"];
	        this.NotSupported = source["NotSupported"];
	        this.Warning = source["Warning"];
	        this.DangerousOperation = source["DangerousOperation"];
	        this.FlashingWrongFirmware = source["FlashingWrongFirmware"];
	        this.AboutToFlashCustom = source["AboutToFlashCustom"];
	        this.AboutToFlashModelPrefix = source["AboutToFlashModelPrefix"];
	        this.AboutToFlashModelSuffix = source["AboutToFlashModelSuffix"];
	        this.MakeSureMatches = source["MakeSureMatches"];
	        this.TypePrefix = source["TypePrefix"];
	        this.TypeSuffix = source["TypeSuffix"];
	        this.PleaseWaitPrefix = source["PleaseWaitPrefix"];
	        this.PleaseWaitSuffix = source["PleaseWaitSuffix"];
	        this.Second = source["Second"];
	        this.Seconds = source["Seconds"];
	        this.TextDoesntMatch = source["TextDoesntMatch"];
	        this.ReadyToProceed = source["ReadyToProceed"];
	        this.FlashNow = source["FlashNow"];
	        this.CustomFirmwareConfirm = source["CustomFirmwareConfirm"];
	        this.USBPermissionsRequired = source["USBPermissionsRequired"];
	        this.CopyRules = source["CopyRules"];
	        this.ThenPaste = source["ThenPaste"];
	        this.ContinueAnyway = source["ContinueAnyway"];
	        this.About = source["About"];
	        this.Copyright = source["Copyright"];
	        this.BuiltWith = source["BuiltWith"];
	        this.Links = source["Links"];
	        this.GitHubRepository = source["GitHubRepository"];
	        this.ReportIssue = source["ReportIssue"];
	        this.ViewLicense = source["ViewLicense"];
	        this.ConsoleOutput = source["ConsoleOutput"];
	        this.Copy = source["Copy"];
	        this.NoLogsAvailable = source["NoLogsAvailable"];
	        this.USBPermissionError = source["USBPermissionError"];
	        this.InstallUdevRules = source["InstallUdevRules"];
	        this.UdevRulesRequired = source["UdevRulesRequired"];
	        this.NoDeviceDetected = source["NoDeviceDetected"];
	        this.FlashOperationFailed = source["FlashOperationFailed"];
	        this.CheckLogsForDetails = source["CheckLogsForDetails"];
	        this.PermissionDenied = source["PermissionDenied"];
	        this.NotCurrentlySupported = source["NotCurrentlySupported"];
	    }
	}

}

