import { useEffect, useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { verifyEmail } from '../api/auth';
import { CheckCircle, XCircle } from 'lucide-react';
import { toast } from '../hooks/use-toast';

const VerifyEmailPage = () => {
  const [searchParams] = useSearchParams();
  const [verificationStatus, setVerificationStatus] = useState<'pending' | 'success' | 'failed'>('pending');
  const [message, setMessage] = useState('');

  useEffect(() => {
    const token = searchParams.get('token');
    if (token) {
      verifyEmail(token)
        .then(() => {
          setVerificationStatus('success');
          setMessage('Your email has been successfully verified! You can now log in.');
          toast({ title: 'Email Verified', description: 'Your email is now verified.' });
        })
        .catch((error: any) => {
          setVerificationStatus('failed');
          setMessage(error.response?.data?.error || 'Email verification failed. The link might be invalid or expired.');
          console.error('Email verification error:', error);
          toast({ title: 'Verification Failed', description: error.response?.data?.error || 'Email verification failed.', variant: 'destructive' });
        });
    } else {
      setVerificationStatus('failed');
      setMessage('No verification token found in the URL.');
      toast({ title: 'Verification Failed', description: 'No token provided.', variant: 'destructive' });
    }
  }, [searchParams]);

  return (
    <div className="flex items-center justify-center min-h-screen bg-background-secondary">
      <div className="w-full max-w-md p-8 space-y-8 bg-background-primary rounded-lg shadow-md text-center">
        {verificationStatus === 'pending' && (
          <>
            <h2 className="text-2xl font-bold font-heading text-text-primary">Verifying your email...</h2>
            <p className="text-text-secondary">Please wait while we process your request.</p>
          </>
        )}
        {verificationStatus === 'success' && (
          <>
            <CheckCircle className="w-20 h-20 mx-auto text-status-success" />
            <h2 className="text-2xl font-bold font-heading text-status-success">Success!</h2>
            <p className="text-text-primary">{message}</p>
            <Link to="/login" className="font-medium text-primary-accent hover:text-secondary-accent">
              Go to Login
            </Link>
          </>
        )}
        {verificationStatus === 'failed' && (
          <>
            <XCircle className="w-20 h-20 mx-auto text-status-error" />
            <h2 className="text-2xl font-bold font-heading text-status-error">Verification Failed</h2>
            <p className="text-text-primary">{message}</p>
            <Link to="/login" className="font-medium text-primary-accent hover:text-secondary-accent">
              Back to Login
            </Link>
          </>
        )}
      </div>
    </div>
  );
};

export default VerifyEmailPage;
