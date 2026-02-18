import { Redirect } from 'expo-router';

// This tab redirects to the gifts index page
export default function GiftsTab() {
  return <Redirect href="/gifts" />;
}
