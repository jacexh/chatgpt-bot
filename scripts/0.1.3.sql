alter table chat
    change current_prompt current text not null;

alter table chat
    drop column channel_internal_id;

alter table conversation
    change answer completion text null;

alter table conversation
    add channel_message_id varchar(48) not null after completion;