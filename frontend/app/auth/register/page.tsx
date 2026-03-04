'use client';

import Link from 'next/link';
import { MobileRedirect } from '@/features/mobile-redirect';
import {
  MOBILE_APP_REDIRECT_PATHS,
  MOBILE_APP_URLS,
} from '@/shared/config/domains';
import { Button } from '@/shared/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

export default function RegisterPage() {
  return (
    <MobileRedirect
      redirectPath={MOBILE_APP_REDIRECT_PATHS.AUTH_REGISTER}
      fallbackUrl={MOBILE_APP_URLS.REGISTER}
    >
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Account Registration Required</CardTitle>
            <CardDescription>
              Registration is handled through our mobile app for enhanced
              security
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Redirecting to mobile app...
            </p>

            <div className="space-y-2">
              <Button asChild variant="outline" className="w-full">
                <Link href={MOBILE_APP_URLS.REGISTER}>
                  Open Mobile Web Version
                </Link>
              </Button>
            </div>
          </CardContent>
          <CardFooter>
            <p className="text-sm text-muted-foreground">
              Already have an account?{' '}
              <Link
                href="/auth/login"
                className="font-medium text-primary hover:underline"
              >
                Sign in
              </Link>
            </p>
          </CardFooter>
        </Card>
      </div>
    </MobileRedirect>
  );
}
