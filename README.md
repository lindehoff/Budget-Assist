# Budget Assist 💰

Ta kontroll över din ekonomi med Budget Assist – din smarta ekonomiassistent som hjälper dig hålla koll på utgifter, sparmål och budget! 🎯

## Vad är Budget Assist? 🤔

Budget Assist är en modern app som kombinerar AI-teknologi med användarvänlig design för att göra privatekonomi enkelt och roligt. Appen läser automatiskt in dina transaktioner, kvitton och räkningar för att ge dig en tydlig bild av din ekonomi.

## Huvudfunktioner ⭐

- 🤖 **Smart Kategorisering** - AI som automatiskt sorterar dina utgifter
- 📊 **Tydlig Översikt** - Se direkt var pengarna tar vägen
- 🎯 **Personliga Budgetar** - Sätt upp och följ dina ekonomiska mål
- 🏠 **Flera Fastigheter & Fordon** - Håll koll på alla dina tillgångar
- 🔒 **Säker & Pålitlig** - Svensk säkerhetsstandard med BankID

## Dokumentation 📚

### För Användare
- [Installationsguide](docs/installation.md) - Kom igång med Budget Assist
- [Konfigurationsguide](docs/configuration.md) - Anpassa appen efter dina behov
- [CLI-kommandon](docs/cli-commands.md) - Fullständig guide till kommandoradsverktyget

### För Utvecklare
- [Utvecklarguide](docs/development-guide.md) - Guide för utvecklare
- [Bidragsguide](docs/CONTRIBUTING.md) - Hur du kan bidra till projektet

### Teknisk Design
- [Systemöversikt](docs/design/00-Overview.md)
- [Systemarkitektur](docs/design/01-System-Architecture.md)
- [Datamodell](docs/design/02-Data-Model.md)
- [AI-Integration](docs/design/03-AI-Integration.md)
- [API-Design](docs/design/04-API-Design.md)
- [Säkerhet & Efterlevnad](docs/design/05-Security-Compliance.md)
- [Utveckling & Driftsättning](docs/design/06-Development-Deployment.md)

### Snabbstart 🚀

1. **Installation**:
   ```bash
   # Med Homebrew (macOS)
   brew install budget-assist

   # Med apt (Ubuntu/Debian)
   sudo apt install budget-assist

   # Manuell installation
   curl -sSL https://get.budgetassist.app | sh
   ```

2. **Konfiguration**:
   ```bash
   # Initiera konfiguration
   budget-assist init

   # Konfigurera databas
   budget-assist config set database.path ~/.budgetassist/data.db
   ```

3. **Första användning**:
   ```bash
   # Importera transaktioner
   budget-assist import file min-kontoutdrag.pdf

   # Visa översikt
   budget-assist transactions list
   ```

## Kom Igång 🚀

1. Ladda ner appen från App Store eller Google Play
2. Logga in med BankID
3. Anslut dina bankkonton eller ladda upp dokument
4. Låt Budget Assist hjälpa dig ta kontroll över din ekonomi!

## Bidra till Utvecklingen 👩‍💻

Vi välkomnar bidrag från community! Kolla in vår [utvecklardokumentation](docs/development-guide.md) och [bidragsguide](docs/CONTRIBUTING.md) för att komma igång.

## Säkerhet & Integritet 🔐

Din data är säker hos oss. Vi använder svensk säkerhetsstandard och följer GDPR. Läs mer i vår [säkerhetsdokumentation](docs/design/05-Security-Compliance.md).

## Få Hjälp 💡

- 📖 Läs vår [dokumentation](docs/)
- 💬 Ställ frågor i [GitHub Discussions](https://github.com/yourusername/Budget-Assist/discussions)
- 🐛 Rapportera problem i [Issue Tracker](https://github.com/yourusername/Budget-Assist/issues)
- 📧 Kontakta support på support@budgetassist.app

---

Utvecklad med ❤️ i Sverige 🇸🇪