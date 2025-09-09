import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { requestPasswordReset } from '../api/auth';
import { toast } from '../hooks/use-toast';
import { useState } from 'react';

const requestResetSchema = z.object({
  email: z.string().email('Invalid email address'),
});

type RequestResetFormData = z.infer<typeof requestResetSchema>;

const RequestPasswordResetPage = () => {
  const [isLoading, setIsLoading] = useState(false);
  const { register, handleSubmit, formState: { errors } } = useForm<RequestResetFormData>({
    resolver: zodResolver(requestResetSchema),
  });
  const navigate = useNavigate();

  const onSubmit = async (data: RequestResetFormData) => {
    setIsLoading(true);
    try {
      await requestPasswordReset(data.email);
      toast({ title: "OTP Sent", description: "A password reset OTP has been sent to your email address." });
      navigate('/reset-password', { state: { email: data.email } });
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || "Failed to send OTP. Please try again.";
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
            Request Password Reset
          </h2>
          <p className="mt-2 text-sm text-text-secondary">
            Enter your email address to receive a password reset OTP.
          </p>
        </div>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          <Input
            id="email"
            type="email"
            placeholder="Email address"
            {...register('email')}
            error={errors.email?.message}
          />
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? 'Sending OTP...' : 'Send OTP'}
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

export default RequestPasswordResetPage;
