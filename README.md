# DesignedbyGG Updater

A modern GUI firmware flasher for DesignedbyGG keyboards powered by SonixFlasher.

## Features

- 🎯 Auto-detect supported keyboards (both application and bootloader modes)
- 🖼️ Visual keyboard identification with images
- 🔒 Safety confirmation dialogs to prevent bricking
- 📝 Real-time console logging
- 🐧 Cross-platform support (Windows, macOS, Linux)
- 🔌 Embedded firmware for all supported models

## Supported Keyboards

- **BSK01** - DesignedbyGG Berserker
- **ICL01** - DesignedbyGG Ironclad v1
- **ICL03** - DesignedbyGG Ironclad v3
- **RB01** - DesignedbyGG RedBladeFS

## Installation

Download the latest release for your platform from the [Releases]() page.

### Linux Requirements

On Linux, udev rules are required for USB access:

1. Copy the udev rules (available in the app via "Copy rules" button):
   ```bash
   sudo tee /etc/udev/rules.d/50-sonix-keyboards.rules