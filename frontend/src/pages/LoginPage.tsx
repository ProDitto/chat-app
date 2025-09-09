import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { useAuthStore } from '../store/authStore';
import { toast } from '../hooks/use-toast';
import { useState } from 'react';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(1, 'Password is required'),
});

type LoginFormData = z.infer<typeof loginSchema>;

const LoginPage = () => {
  const [isLoading, setIsLoading] = useState(false);
  const { register, handleSubmit, formState: { errors } } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });
  const navigate = useNavigate();
  const { login } = useAuthStore();

  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true);
    try {
      await login(data.email, data.password);
      toast({ title: "Login successful!" });
      navigate('/');
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || "Please check your credentials.";
      toast({ title: "Login failed", description: errorMessage, variant: "destructive" });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-background-secondary">
      <div className="w-full max-w-md p-8 space-y-8 bg-background-primary rounded-lg shadow-md">
        <div className="text-center">
          <h2 className="mt-6 text-3xl font-bold font-heading text-text-primary">
            Sign in to your account
          </h2>
        </div>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <Input
                id="email"
                type="email"
                placeholder="Email address"
                {...register('email')}
                error={errors.email?.message}
                className="mb-2"
              />
            </div>
            <div>
              <Input
                id="password"
                type="password"
                placeholder="Password"
                {...register('password')}
                error={errors.password?.message}
              />
            </div>
          </div>
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? 'Signing in...' : 'Sign in'}
          </Button>
        </form>
        <p className="text-sm text-center text-text-secondary">
          <Link to="/request-password-reset" className="font-medium text-primary-accent hover:text-secondary-accent">
            Forgot password?
          </Link>
        </p>
        <p className="text-sm text-center text-text-secondary">
          Don't have an account?{' '}
          <Link to="/register" className="font-medium text-primary-accent hover:text-secondary-accent">
            Sign up
          </Link>
        </p>
      </div>
    </div>
  );
};

export default LoginPage;
