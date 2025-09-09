import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
import { registerUser } from '../api/auth';
import { toast } from '../hooks/use-toast';
import { useState } from 'react';

const registerSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
});

type RegisterFormData = z.infer<typeof registerSchema>;

const RegisterPage = () => {
  const [isLoading, setIsLoading] = useState(false);
  const { register, handleSubmit, formState: { errors } } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  });
  const navigate = useNavigate();

  const onSubmit = async (data: RegisterFormData) => {
    setIsLoading(true);
    try {
      await registerUser(data);
      toast({ title: "Registration successful!", description: "Please check your email to verify your account." });
      navigate('/login');
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || "This email or username may already be in use.";
      toast({ title: "Registration failed", description: errorMessage, variant: "destructive" });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-background-secondary">
      <div className="w-full max-w-md p-8 space-y-8 bg-background-primary rounded-lg shadow-md">
        <div className="text-center">
          <h2 className="mt-6 text-3xl font-bold font-heading text-text-primary">
            Create an account
          </h2>
        </div>
        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <Input id="username" type="text" placeholder="Username" {...register('username')} error={errors.username?.message} className="mb-2" />
            </div>
            <div>
              <Input id="email" type="email" placeholder="Email address" {...register('email')} error={errors.email?.message} className="mb-2" />
            </div>
            <div>
              <Input id="password" type="password" placeholder="Password" {...register('password')} error={errors.password?.message} />
            </div>
          </div>
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? 'Creating account...' : 'Sign up'}
          </Button>
        </form>
         <p className="text-sm text-center text-text-secondary">
          Already have an account?{' '}
          <Link to="/login" className="font-medium text-primary-accent hover:text-secondary-accent">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
};

export default RegisterPage;
