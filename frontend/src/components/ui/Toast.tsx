// This is a complex component, often taken from a library like shadcn/ui.
// For brevity, a simplified version is provided. Full implementation would handle animations and variants.
import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { X } from "lucide-react"

import { cn } from "../../lib/utils"

type ToastProps = {
    variant?: string
    open?: Boolean
    onOpenChange?: (isOpen: Boolean) => void
    duration?: number
}

const ToastProvider = ({ children }: { children: React.ReactNode }) => <div>{children}</div>
const ToastViewport = () => <div className="fixed top-0 z-[100] flex max-h-screen w-full flex-col-reverse p-4 sm:bottom-0 sm:right-0 sm:top-auto sm:flex-col md:max-w-[420px]" />

const toastVariants = cva(
    "group pointer-events-auto relative flex w-full items-center justify-between space-x-4 overflow-hidden rounded-md border p-6 pr-8 shadow-lg transition-all",
    {
        variants: {
            variant: {
                default: "border bg-background-primary text-text-primary",
                destructive: "destructive group border-status-error bg-status-error text-white",
            },
        },
        defaultVariants: {
            variant: "default",
        },
    }
)

const Toast = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & VariantProps<typeof toastVariants>>(({ className, variant, ...props }, ref) => {
    return (
        <div ref={ref} className={cn(toastVariants({ variant }), className)} {...props} />
    )
})
Toast.displayName = "Toast"

const ToastClose = React.forwardRef<HTMLButtonElement, React.ButtonHTMLAttributes<HTMLButtonElement>>(({ className, ...props }, ref) => (
    <button ref={ref} type="button" className={cn("absolute right-2 top-2 rounded-md p-1 text-inherit opacity-70 transition-opacity hover:opacity-100 focus:opacity-100 focus:outline-none focus:ring-2 group-hover:opacity-100", className)} {...props}>
        <X className="h-4 w-4" />
    </button>
))
ToastClose.displayName = "ToastClose"

const ToastTitle = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(({ className, ...props }, ref) => (
    <p ref={ref} className={cn("text-sm font-semibold", className)} {...props} />
))
ToastTitle.displayName = "ToastTitle"

const ToastDescription = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(({ className, ...props }, ref) => (
    <p ref={ref} className={cn("text-sm opacity-90", className)} {...props} />
))
ToastDescription.displayName = "ToastDescription"

export { Toast, ToastProvider, ToastViewport, ToastTitle, ToastDescription, ToastClose, type ToastProps };

