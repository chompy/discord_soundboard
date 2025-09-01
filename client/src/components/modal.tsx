import React from 'react';

export type ModalProperties = {
    children: React.JSX.Element;
    isOpen: boolean;
    close?: () => void;
};

function Modal({ children, isOpen, close }: ModalProperties) {
    if (!isOpen) return;

    const onClickOutsideClose = (e: object) => {
        if (
            'target' in e &&
            e.target &&
            typeof e.target === 'object' &&
            'className' in e.target &&
            e.target.className == 'modal'
        ) {
            close?.();
        }
    };

    return (
        <>
            <div className="modal" onClick={onClickOutsideClose}>
                <div className="modal-inner">{children}</div>
            </div>
        </>
    );
}

export default Modal;
