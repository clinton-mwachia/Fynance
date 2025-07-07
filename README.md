## Packaging the Application

```bash
fyne package -os windows
```

### Additional Packaging Options

The `fyne` package command can be enhanced with various flags for more customization:

1. `-icon Icon.png`: Specify a different icon file.
2. `-appID <id>`: Override the app ID from FyneApp.toml.
3. `-name <name>`: Override the app name.
4. `-release`: Build a release version with optimizations.
5. `-executable <file>`: Specify the executable name if different from the default.

### Packaging for Different Platforms

#### **Windows (.exe Installer)**

1. **Download and Install Inno Setup**: [jrsoftware.org](https://jrsoftware.org/isinfo.php)
2. **Create a New Script**: Use Inno Setup to create a new installer script.
3. **Add Application Files**: Include your compiled `.exe` and other required files.
4. **Compile the Installer**: Follow the wizard to generate an `.exe` installer.

Example Inno Setup Script:

```ini
[Setup]
AppName=Fynance
AppVersion=1.0
DefaultDirName={pf}\fynance
DefaultGroupName=fynance
OutputDir=output
OutputBaseFilename=fynanceSetup
Compression=lzma
SolidCompression=yes

[Files]
Source: "fynance.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\*"; DestDir: "{app}\assets"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\Fynance"; Filename: "{app}\fynance.exe"
Name: "{group}\Uninstall Fynance"; Filename: "{uninstallexe}"
```

## Making the Application Installable

Once you have packaged the application for your target platform, distribute the installer file (`.exe`, `.dmg`, `.deb`, etc.) to users. They can then run the installer to install your Fynancelication on their system.

## Running the Application

After installation, you can run the application from the system's application menu or by executing the installed binary directly.

```bash
./GoDesktopApp   # Linux/macOS
GoDesktopApp.exe # Windows
```

## License

This project is licensed under the Appache-2.0 License - see the [LICENSE](LICENSE) file for details.
