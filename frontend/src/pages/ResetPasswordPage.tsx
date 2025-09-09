import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { resetPassword } from '../api/auth';
import { toast } from '../hooks/use-toast';
import { useState } from 'react';

const resetPasswordSchema = z.object({
  otp: z.string().min(6, 'OTP must be 6 digits').max(6, 'OTP must be 6 digits'),
  newPassword: z.string().min(8, 'Password must be at least 8 characters'),
});

type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>;

const ResetPasswordPage = () => {
  const [isLoading, setIsLoading] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const email = location.state?.email || ''; // Pre-fill email from request page

  const { register, handleSubmit, formState: { errors } } = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema),
  });

  const onSubmit = async (data: ResetPasswordFormData) => {
    if (!email) {
      toast({ title: "Error", description: "Email not provided. Please go back and request an OTP.", variant: "destructive" });
      return;
    }
    setIsLoading(true);
    try {
      await resetPassword(email, data.otp, data.newPassword);
      toast({ title: "Password Reset", description: "Your password has been successfully reset. You can now log in with your new password." });
      navigate('/login');
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || "Failed to reset password. Check your OTP and try again.";
      toast({ title: "Error", description: errorMessage, variant: "destructive" });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-background-secondary">
      <div className="w-full max-w-md p-8 space-y-8 bg-background-primary rounded-lg shadow-md">
        <div className="text-center">
          <h2 className="mt-6 text-2xl font-bold font-heading text-text-primary">
            Reset Your Password
          </h2>
          <p className="mt-2 text-sm text-text-secondary">
            Enter the OTP sent to <span className="font-semibold">{email}</span> and your new password.
          </p>
        </div>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          <Input
            id="otp"
            type="text"
            placeholder="One-Time Password (OTP)"
            {...register('otp')}
            error={errors.otp?.message}
            maxLength={6}
            className="mb-2"
          />
          <Input
            id="newPassword"
            type="password"
            placeholder="New Password"
            {...register('newPassword')}
            error={errors.newPassword?.message}
          />
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? 'Resetting Password...' : 'Reset Password'}
          </Button>
        </form>
        <p className="text-sm text-center text-text-secondary">
          <Link to="/login" className="font-medium text-primary-accent hover:text-secondary-accent">
            Back to Login
          </Link>
        </p>
      </div>
    </div>
  );
};

export default ResetPasswordPage;
