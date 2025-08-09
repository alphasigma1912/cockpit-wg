import { mockBackend } from '../support/stub';

describe('App flows', () => {
  it('handles first run', () => {
    cy.visit('/', {
      onBeforeLoad: (win) => mockBackend(win, { CheckPrereqs: { kernel: false, tools: false, systemd: false } }),
    });
    cy.contains('WireGuard is not installed');
    cy.contains('Install WireGuard').click();
    cy.contains('Installation complete');
    cy.injectAxe();
    cy.checkA11y();
  });

  it('adds a peer', () => {
    cy.visit('/', {
      onBeforeLoad: (win) => mockBackend(win),
    });
    cy.contains('Peers').click();
    cy.get('#endpoint').type('1.2.3.4');
    cy.get('#allowed').type('10.0.0.0/24');
    cy.contains('Add peer').click();
    cy.contains('Public key');
    cy.injectAxe();
    cy.checkA11y();
  });

  it('edits interface and views traffic', () => {
    cy.visit('/', {
      onBeforeLoad: (win) => mockBackend(win),
    });
    cy.contains('Interfaces').click();
    cy.contains('Down').click();
    cy.contains('Down').last().click();
    cy.get('svg').should('exist');
    cy.injectAxe();
    cy.checkA11y();
  });

  it('imports bundle', () => {
    cy.visit('/', {
      onBeforeLoad: (win) => mockBackend(win),
    });
    cy.contains('Exchange').click();
    cy.get('[data-testid="bundle-input"]').selectFile('cypress/fixtures/bundle.wgx');
    cy.contains('Bundle imported');
    cy.injectAxe();
    cy.checkA11y();
  });
});
