CREATE TABLE IF NOT EXISTS public.votes
(
    id uuid NOT NULL,
    target text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT "Votes_pkey" PRIMARY KEY (id)
)
