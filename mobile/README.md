# Wish List Mobile App

This is an [Expo](https://expo.dev) project created with [`create-expo-app`](https://www.npmjs.com/package/create-expo-app). It serves as the mobile application for the Wish List platform.

## Features

- **User Authentication**: Email/password and OAuth with Google, Facebook, and Apple
- **Wish List Management**: Create, edit, and share wish lists
- **Responsive Design**: Works seamlessly on iOS and Android
- **Dark/Light Theme**: Automatic theme switching based on system preference
- **Modern UI**: Built with react-native-paper for consistent Material Design

## Get started

1. Install dependencies

   ```bash
   npm install
   ```

2. Configure environment variables (see [.env.example](./.env.example))

   ```bash
   cp .env.example .env
   # Edit .env with your specific values
   ```

3. Start the app

   ```bash
   npx expo start
   ```

In the output, you'll find options to open the app in a

- [development build](https://docs.expo.dev/develop/development-builds/introduction/)
- [Android emulator](https://docs.expo.dev/workflow/android-studio-emulator/)
- [iOS simulator](https://docs.expo.dev/workflow/ios-simulator/)
- [Expo Go](https://expo.dev/go), a limited sandbox for trying out app development with Expo

You can start developing by editing the files inside the **app** directory. This project uses [file-based routing](https://docs.expo.dev/router/introduction).

## Authentication

The app supports multiple authentication methods:
- Traditional email/password
- OAuth with Google, Facebook, and Apple

For detailed information about authentication setup, see [docs/AUTHENTICATION.md](./docs/AUTHENTICATION.md).

## Get a fresh project

When you're ready, run:

```bash
npm run reset-project
```

This command will move the starter code to the **app-example** directory and create a blank **app** directory where you can start developing.

## Learn more

To learn more about developing your project with Expo, look at the following resources:

- [Expo documentation](https://docs.expo.dev/): Learn fundamentals, or go into advanced topics with our [guides](https://docs.expo.dev/guides).
- [Learn Expo tutorial](https://docs.expo.dev/tutorial/introduction/): Follow a step-by-step tutorial where you'll create a project that runs on Android, iOS, and the web.

## Join the community

Join our community of developers creating universal apps.

- [Expo on GitHub](https://github.com/expo/expo): View our open source platform and contribute.
- [Discord community](https://chat.expo.dev): Chat with Expo users and ask questions.
