import { useEffect, useRef } from 'react';
import { fetchEvents } from '../api/events';
import { toast } from './use-toast';
import { type Event } from '../types/event'; // Import Event type

const LONG_POLLING_INTERVAL_MS = 5000; // 5 seconds

export const useLongPolling = (accessToken: string | null, isWsConnected: boolean, onEvent: (event: Event) => void) => {
  const intervalRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastEventIdRef = useRef<string | undefined>(undefined); // Now tracks event ID

  const poll = async () => {
    if (!accessToken || isWsConnected) {
      // Don't poll if not authenticated or WebSocket is connected
      if (intervalRef.current) {
        clearTimeout(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

    try {
      const events = await fetchEvents(lastEventIdRef.current); // Pass lastEventId
      if (events && events.length > 0) {
        events.forEach(event => onEvent(event));
        // Update lastEventId to the ID of the latest event
        lastEventIdRef.current = events[events.length - 1].id; 
      }
    } catch (error: any) {
      if (error.code === 'ECONNABORTED' || error.response?.status === 408) {
        // Request timed out (expected for long polling, not an error)
        // console.log("Long polling request timed out, re-polling...");
      } else {
        console.error("Long polling error:", error);
        toast({ title: "Long Polling Error", description: "Failed to fetch events.", variant: "destructive" });
        // Consider exponential backoff here if needed, or simply retry after interval
      }
    } finally {
      // Schedule the next poll regardless of success or failure
      if (intervalRef.current) {
        clearTimeout(intervalRef.current);
      }
      intervalRef.current = setTimeout(poll, LONG_POLLING_INTERVAL_MS);
    }
  };

  useEffect(() => {
    if (accessToken && !isWsConnected) {
      // Start polling only if authenticated and WS is NOT connected
      if (!intervalRef.current) {
        console.log("Starting long polling...");
        lastEventIdRef.current = undefined; // Reset lastEventId when starting polling
        poll();
      }
    } else {
      // Stop polling if not authenticated or WS is connected
      if (intervalRef.current) {
        console.log("Stopping long polling...");
        clearTimeout(intervalRef.current);
        intervalRef.current = null;
      }
    }

    // Cleanup on component unmount
    return () => {
      if (intervalRef.current) {
        clearTimeout(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [accessToken, isWsConnected, onEvent]);
};