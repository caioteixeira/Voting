CREATE TABLE IF NOT EXISTS public."Votes"
(
    id uuid NOT NULL,
    "Target" text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT "Votes_pkey" PRIMARY KEY (id)
)
