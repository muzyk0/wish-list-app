'use client';

import Link from 'next/link';
import MobileRedirect from '@/components/common/MobileRedirect';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  MOBILE_APP_REDIRECT_PATHS,
  MOBILE_APP_URLS,
} from '@/constants/domains';

export default function LoginPage() {
  return (
    <MobileRedirect
      redirectPath={MOBILE_APP_REDIRECT_PATHS.AUTH_LOGIN}
      fallbackUrl={MOBILE_APP_URLS.LOGIN}
    >
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Account Access Required</CardTitle>
            <CardDescription>
              Login and registration are handled through our mobile app for
              enhanced security
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Redirecting to mobile app...
            </p>

            <div className="space-y-2">
              <Button asChild variant="outline" className="w-full">
                <Link href={MOBILE_APP_URLS.LOGIN}>
                  Open Mobile Web Version
                </Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </MobileRedirect>
  );
}
